import { useCallback } from 'react'
import { useChunkUpload } from './useChunkUpload'
import type { UploadProgress } from './useChunkUpload'
import * as hfsApi from '@/api/hfs'

/**
 * 头像 / 图标等静态资源上传 Hook。
 *
 * 流程：分片上传 → merge → POST /api/hfs/assets → 返回可访问路径。
 * 调用方拿到 path 后自行更新到对应字段（如用户头像）。
 */
export function useAvatarUpload() {
  const { upload: chunkUpload, progress, error, cancel, isUploading } = useChunkUpload()

  const upload = useCallback(
    async (file: File): Promise<string> => {
      const mergeResult = await chunkUpload(file)
      const { path } = await hfsApi.commitAssets(mergeResult.uploadId)
      return path
    },
    [chunkUpload],
  )

  return { upload, progress, error, cancel, isUploading } as {
    upload: (file: File) => Promise<string>
    progress: UploadProgress | null
    error: Error | null
    cancel: () => void
    isUploading: boolean
  }
}
