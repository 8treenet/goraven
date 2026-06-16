import { useCallback } from 'react'
import { useChunkUpload } from './useChunkUpload'
import type { UploadProgress } from './useChunkUpload'
import { commitUpload } from '@/api/files'

/**
 * 文件管理器上传 Hook。
 *
 * 流程：分片上传 → merge → POST /api/fileManager/upload → 文件移入用户空间。
 * 返回文件在用户空间中的相对路径。
 */
export function useFileUpload() {
  const { upload: chunkUpload, progress, error, cancel, isUploading } = useChunkUpload()

  const upload = useCallback(
    async (file: File, targetDir?: string): Promise<string> => {
      const mergeResult = await chunkUpload(file)
      const { path } = await commitUpload(mergeResult.uploadId, targetDir)
      return path
    },
    [chunkUpload],
  )

  return { upload, progress, error, cancel, isUploading } as {
    upload: (file: File, targetDir?: string) => Promise<string>
    progress: UploadProgress | null
    error: Error | null
    cancel: () => void
    isUploading: boolean
  }
}
