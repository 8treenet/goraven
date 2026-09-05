import { useCallback, useMemo, forwardRef } from 'react'
import { useT } from '@/i18n'
import {
  listTeamFiles,
  commitTeamUpload,
  teamMkdir,
  teamRename,
  teamDeleteFiles,
  teamCompress,
  teamDecompress,
  getTeamDownloadUrl,
  createTeamTempAccess,
} from '@/api/team-projects'
import type { TeamProjectItem } from '@/api/types'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { FileList, type FileListApi, type FileListHandle } from './FileList'

interface TeamProjectFilesProps {
  project: TeamProjectItem
  onBack: () => void
}

export type TeamProjectFilesHandle = FileListHandle

export const TeamProjectFiles = forwardRef<TeamProjectFilesHandle, TeamProjectFilesProps>(function TeamProjectFiles({ project, onBack }, ref) {
  const t = useT()
  const { upload: chunkUpload, progress: uploadProgress, isUploading } = useChunkUpload()

  const api = useMemo<FileListApi>(() => ({
    list: (dir) => listTeamFiles(project.id, dir),
    mkdir: (dir) => teamMkdir(project.id, dir),
    rename: (oldPath, newPath) => teamRename(project.id, oldPath, newPath),
    remove: (paths) => teamDeleteFiles(project.id, paths),
    compress: (req) => teamCompress(project.id, req),
    decompress: (req) => teamDecompress(project.id, req),
    downloadUrl: (path) => getTeamDownloadUrl(project.id, path),
    createAccess: (path, type) => createTeamTempAccess(project.id, path, type),
  }), [project.id])

  const projectRootPath = useMemo(() => `projects/${project.projectName}`, [project.projectName])

  const buildAkPath = useCallback((filePath: string) => `${projectRootPath}${filePath}`, [projectRootPath])

  const uploadFile = useCallback(async (file: File, dir: string) => {
    const mergeResult = await chunkUpload(file)
    await commitTeamUpload(project.id, mergeResult.uploadId, dir)
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
      errorTitle={t('files.projectNotFound')}
      errorActionLabel={t('files.teamProjects')}
      errorAction={onBack}
    />
  )
})
