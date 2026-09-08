import { useCallback, useMemo, forwardRef } from 'react'
import { useT } from '@/i18n'
import {
  listMyFiles,
  commitMyUpload,
  myMkdir,
  myRename,
  myDeleteFiles,
  myCompress,
  myDecompress,
  getMyDownloadUrl,
  createMyTempAccess,
} from '@/api/my-projects'
import type { MyProjectItem } from '@/api/types'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { FileList, type FileListApi, type FileListHandle } from './FileList'

interface MyProjectFilesProps {
  project: MyProjectItem
  onBack: () => void
}

export type MyProjectFilesHandle = FileListHandle

export const MyProjectFiles = forwardRef<MyProjectFilesHandle, MyProjectFilesProps>(function MyProjectFiles({ project, onBack }, ref) {
  const t = useT()
  const { upload: chunkUpload, progress: uploadProgress, isUploading } = useChunkUpload()

  const api = useMemo<FileListApi>(() => ({
    list: (dir) => listMyFiles(project.id, dir),
    mkdir: (dir) => myMkdir(project.id, dir),
    rename: (oldPath, newPath) => myRename(project.id, oldPath, newPath),
    remove: (paths) => myDeleteFiles(project.id, paths),
    compress: (req) => myCompress(project.id, req),
    decompress: (req) => myDecompress(project.id, req),
    downloadUrl: (path) => getMyDownloadUrl(project.id, path),
    createAccess: (path, type) => createMyTempAccess(project.id, path, type),
  }), [project.id])

  const projectRootPath = useMemo(() => `projects/${project.projectName}`, [project.projectName])

  const buildAkPath = useCallback((filePath: string) => `${projectRootPath}${filePath}`, [projectRootPath])

  const uploadFile = useCallback(async (file: File, dir: string) => {
    const mergeResult = await chunkUpload(file)
    await commitMyUpload(project.id, mergeResult.uploadId, dir)
  }, [chunkUpload, project.id])

  return (
    <FileList
      ref={ref}
      initialDir="/"
      rootLabel={project.projectName}
      onBackAtRoot={onBack}
      api={api}
      uploadFile={uploadFile}
      uploadProgress={uploadProgress}
      isUploading={isUploading}
      buildAkPath={buildAkPath}
      errorTitle={t('files.myProjectNotFound')}
      errorActionLabel={t('files.myProjects')}
      errorAction={onBack}
    />
  )
})