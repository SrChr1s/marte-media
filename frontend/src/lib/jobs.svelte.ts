import { EventsOn } from '../../wailsjs/runtime/runtime'
import { ListJobs, OpenFile } from '../../wailsjs/go/main/App'
import { toast } from './toasts.svelte'
import type { Job } from './types'

// Live job list, fed by "job:update" events from the Go queue.
export const jobs = $state<{ list: Job[] }>({ list: [] })

export function activeCount(): number {
  return jobs.list.filter((j) => j.status === 'queued' || j.status === 'downloading' || j.status === 'processing').length
}

function upsert(j: Job) {
  const i = jobs.list.findIndex((x) => x.id === j.id)
  const prev = i >= 0 ? jobs.list[i].status : undefined
  if (i >= 0) jobs.list[i] = j
  else jobs.list.push(j)

  // Tell the user when a job ends, whichever view they are on.
  const title = j.request.title || j.request.url
  if (prev && prev !== j.status) {
    if (j.status === 'done') toast('ok', `Descarga completada: ${title}`, { label: 'Abrir', run: () => OpenFile(j.path) })
    else if (j.status === 'error') toast('error', `No se pudo descargar: ${title}`)
  }
}

export async function initJobs() {
  jobs.list = ((await ListJobs()) as unknown as Job[]) ?? []
  EventsOn('job:update', (j: Job) => upsert(j))
}

// ClearFinished on the Go side drops these; mirror it locally.
export function dropFinished() {
  jobs.list = jobs.list.filter((j) => !['done', 'error', 'canceled'].includes(j.status))
}
