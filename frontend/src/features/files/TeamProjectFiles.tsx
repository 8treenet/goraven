import { useCallback, useEffect, useMemo, useRef, useState, forwardRef } from 'react'
import { Folder, FolderGit2, GitBranch } from 'lucide-react'
import { useT } from '@/i18n'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
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
  teamProjectGitApi,
} from '@/api/team-projects'
import type { TeamProjectItem, GitStatus } from '@/api/types'
import { useChunkUpload } from '@/hooks/useChunkUpload'
import { FileList, type FileListApi, type FileListHandle } from './FileList'
import { GitDialog } from './git/GitDialog'
import { GIT_CLONE_STATE, buildChangeMap } from './git/git-helpers'

interface TeamProjectFilesProps {
  project: TeamProjectItem
  onBack: () => void
}

export type TeamProjectFilesHandle = FileListHandle

export const TeamProjectFiles = forwardRef<TeamProjectFilesHandle, TeamProjectFilesProps>(function TeamProjectFiles({ project, onBack }, ref) {
  const t = useT()
  const { upload: chunkUpload, progress: uploadProgress, isUploading } = useChunkUpload()
  const [gitStatus, setGitStatus] = useState<GitStatus | null>(null)
  const [gitOpen, setGitOpen] = useState(false)
  const fileListRef = useRef<FileListHandle | null>(null)

  const refreshGit = useCallback(() => {
    teamProjectGitApi.getStatus(project.id)
      .then((rsp) => setGitStatus(rsp))
      .catch(() => {})
  }, [project.id])

  useEffect(() => {
    refreshGit()
  }, [refreshGit])

  // 克隆进行中每 3 秒轮询直至终态
  useEffect(() => {
    if (gitStatus?.cloneState !== GIT_CLONE_STATE.RUNNING) return
    const timer = setInterval(refreshGit, 3000)
    return () => clearInterval(timer)
  }, [gitStatus?.cloneState, refreshGit])

  const gitChangeMap = useMemo(
    () => (gitStatus?.enabled ? buildChangeMap(gitStatus.changes) : null),
    [gitStatus],
  )

  const gitStatusOf = useCallback((path: string) => {
    const change = gitChangeMap?.get(path.replace(/^\//, ''))
    return change ? change.status : null
  }, [gitChangeMap])

  // Git 入口按钮：状态着色 + 未提交变更数徽标
  const gitEnabled = !!gitStatus?.enabled
  const gitHasRemote = !!gitStatus?.remote?.url
  const gitChangeCount = gitStatus?.changes.length ?? 0

  const toolbarExtra = (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          onClick={() => setGitOpen(true)}
          className="relative flex items-center gap-1 rounded-md px-2 py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover hover:text-text-1"
        >
          {!gitEnabled ? (
            <Folder className="size-3.5 text-text-3" />
          ) : gitHasRemote ? (
            <GitBranch className="size-3.5 text-success" />
          ) : (
            <FolderGit2 className="size-3.5 text-warning" />
          )}
          {gitEnabled && gitChangeCount > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex h-3.5 min-w-3.5 items-center justify-center rounded-full bg-highlight px-1 text-[9px] font-medium leading-none text-highlight-fg">
              {gitChangeCount}
            </span>
          )}
        </button>
      </TooltipTrigger>
      <TooltipContent>
        {!gitEnabled ? t('files.gitTipNone') : gitHasRemote ? t('files.gitTipOrigin') : t('files.gitTipLocal')}
      </TooltipContent>
    </Tooltip>
  )

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
    <>
      <FileList
        ref={(handle) => {
          fileListRef.current = handle
          if (typeof ref === 'function') ref(handle)
          else if (ref) (ref as React.MutableRefObject<FileListHandle | null>).current = handle
        }}
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
        toolbarExtra={toolbarExtra}
        gitStatusOf={gitStatusOf}
      />
      <GitDialog
        open={gitOpen}
        onOpenChange={setGitOpen}
        projectName={project.projectName}
        projectId={project.id}
        gitApi={teamProjectGitApi}
        canConfig={project.isCreator}
        onMutated={refreshGit}
        onAfterPull={() => fileListRef.current?.refresh()}
      />
    </>
  )
})
