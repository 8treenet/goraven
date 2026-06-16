import { useState, useRef, useCallback } from 'react'
import * as hfsApi from '@/api/hfs'

// ============================================================
// 分片大小 800KB，小于 nginx 默认 client_max_body_size (1MB)
// ============================================================
const DEFAULT_CHUNK_SIZE = 800 * 1024

// ============================================================
// 类型
// ============================================================

/** 上传进度 */
export interface UploadProgress {
  fileName: string
  uploadedChunks: number
  totalChunks: number
  percentage: number
  stage: 'creating' | 'uploading' | 'merging' | 'done'
}

/** merge 成功后 HFS 返回的数据 */
export interface MergeResult {
  uploadId: string
  filePath: string
  fileName: string
  fileSize: number
}

export interface UseChunkUploadOptions {
  /** 每片大小（字节），默认 800KB */
  chunkSize?: number
  /** 进度回调 */
  onProgress?: (progress: UploadProgress) => void
}

export interface UseChunkUploadReturn {
  /**
   * 上传单个文件，返回 merge 结果。
   * 多文件场景由调用方依次调用，不内部并行。
   */
  upload: (file: File) => Promise<MergeResult>
  /** 当前进度 */
  progress: UploadProgress | null
  /** 最近一次错误 */
  error: Error | null
  /** 取消当前上传 */
  cancel: () => void
  /** 是否正在上传 */
  isUploading: boolean
}

// ============================================================
// Hook
// ============================================================

/**
 * 核心分片上传 Hook。
 *
 * 封装 HFS 的 create → chunk(0..N) → merge 三段流程：
 * - 自动分片（默认 800KB）
 * - 逐片顺序上传，同时上报进度
 * - 支持取消（在分片间隙检查）
 * - 不负责 merge 之后的业务操作（调 assets、fileManager 等）
 *
 * 多文件场景：调用方逐个调用 `upload()`，不支持内部并行。
 */
export function useChunkUpload(options: UseChunkUploadOptions = {}): UseChunkUploadReturn {
  const { chunkSize = DEFAULT_CHUNK_SIZE, onProgress } = options

  const [progress, setProgress] = useState<UploadProgress | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  // 取消标记 — 用 ref 避免闭包依赖、避免不必要的 re-render
  const cancelledRef = useRef(false)

  const cancel = useCallback(() => {
    cancelledRef.current = true
  }, [])

  const upload = useCallback(
    async (file: File): Promise<MergeResult> => {
      // ---- 重置状态 ----
      cancelledRef.current = false
      setError(null)
      setIsUploading(true)

      const totalChunks = Math.ceil(file.size / chunkSize)

      // 初始化进度
      const initProgress: UploadProgress = {
        fileName: file.name,
        uploadedChunks: 0,
        totalChunks,
        percentage: 0,
        stage: 'creating',
      }
      setProgress(initProgress)
      onProgress?.(initProgress)

      try {
        // ============ Stage 1: create ============
        const { uploadId } = await hfsApi.createUpload({
          fileName: file.name,
          fileSize: file.size,
          chunkSize,
          totalChunks,
        })

        if (cancelledRef.current) throw CANCELLED_ERROR

        // ============ Stage 2: chunk upload（顺序） ============
        for (let i = 0; i < totalChunks; i++) {
          if (cancelledRef.current) throw CANCELLED_ERROR

          const start = i * chunkSize
          const end = Math.min(start + chunkSize, file.size)
          const blob = file.slice(start, end)

          await hfsApi.uploadChunk(uploadId, i, blob)

          const uploaded = i + 1
          const pct = Math.round((uploaded / totalChunks) * 100)
          const chunkProgress: UploadProgress = {
            fileName: file.name,
            uploadedChunks: uploaded,
            totalChunks,
            percentage: pct,
            stage: 'uploading',
          }
          setProgress(chunkProgress)
          onProgress?.(chunkProgress)
        }

        if (cancelledRef.current) throw CANCELLED_ERROR

        // ============ Stage 3: merge ============
        setProgress((prev) => (prev ? { ...prev, stage: 'merging' } : null))
        onProgress?.({ fileName: file.name, uploadedChunks: totalChunks, totalChunks, percentage: 100, stage: 'merging' })

        const mergeResult = await hfsApi.mergeUpload(uploadId)

        const doneProgress: UploadProgress = {
          fileName: file.name,
          uploadedChunks: totalChunks,
          totalChunks,
          percentage: 100,
          stage: 'done',
        }
        setProgress(doneProgress)
        onProgress?.(doneProgress)

        return {
          uploadId: mergeResult.uploadId,
          filePath: mergeResult.filePath,
          fileName: mergeResult.fileName,
          fileSize: mergeResult.fileSize,
        }
      } catch (err) {
        // 取消不算错误
        if (err === CANCELLED_ERROR) throw err

        const wrapped = err instanceof Error ? err : new Error(String(err))
        setError(wrapped)
        throw wrapped
      } finally {
        setIsUploading(false)
      }
    },
    [chunkSize, onProgress],
  )

  return { upload, progress, error, cancel, isUploading }
}

/** 内部哨兵，用于在 catch 中区分「取消」和「真实错误」 */
const CANCELLED_ERROR = new Error('Upload cancelled')
