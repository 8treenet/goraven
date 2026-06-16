import { useCallback } from 'react'
import { useChunkUpload } from './useChunkUpload'
import type { UploadProgress } from './useChunkUpload'

/**
 * 聊天附件上传 Hook。
 *
 * 流程：分片上传 → merge → 返回 uploadId。
 * 调用方将 uploadId 放入 POST /api/chat 的 attachments 数组。
 *
 * 与 useFileUpload 的区别：
 * - useFileUpload：merge → POST /api/fileManager/upload → 文件移入用户空间
 * - useChatAttachment：merge → 直接返回 uploadId，不做额外操作
 */
export function useChatAttachment() {
  const { upload: chunkUpload, progress, error, cancel, isUploading } = useChunkUpload()

  const upload = useCallback(
    async (file: File): Promise<string> => {
      const mergeResult = await chunkUpload(file)
      return mergeResult.uploadId
    },
    [chunkUpload],
  )

  return { upload, progress, error, cancel, isUploading } as {
    /** 上传文件，返回 uploadId（放入 chat attachments 数组） */
    upload: (file: File) => Promise<string>
    progress: UploadProgress | null
    error: Error | null
    cancel: () => void
    isUploading: boolean
  }
}
