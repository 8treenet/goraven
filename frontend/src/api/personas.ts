import http from './http'
import type {
  PersonaSimple,
  PersonaListItem,
  PersonaDetail,
  PersonaCategory,
  PersonaTemplateItem,
  PersonaTemplateDetail,
  CreatePersonaRequest,
  UpdatePersonaRequest,
} from './types'

/** GET /api/personas */
export function getPersonas() {
  return http.get<PersonaListItem[]>('/personas')
}

/** GET /api/personas?simple=true (sidebar — lightweight) */
export function getPersonasSimple() {
  return http.get<PersonaSimple[]>('/personas', { params: { simple: 'true' } })
}

/** GET /api/personas/:id */
export function getPersonaDetail(id: number) {
  return http.get<PersonaDetail>(`/personas/${id}`)
}

/** POST /api/personas */
export function createPersona(data: CreatePersonaRequest) {
  return http.post<{ status: string }>('/personas', data)
}

/** PUT /api/personas/:id */
export function updatePersona(id: number, data: UpdatePersonaRequest) {
  return http.put<{ status: string }>(`/personas/${id}`, data)
}

/** DELETE /api/personas/:id */
export function deletePersona(id: number) {
  return http.delete(`/personas/${id}`)
}

/** GET /api/personas/personaTemplates */
export function getTemplates() {
  return http.get<PersonaTemplateItem[]>('/personas/personaTemplates')
}

/** GET /api/personas/personaTemplates/:id */
export function getTemplateDetail(id: number) {
  return http.get<PersonaTemplateDetail>(`/personas/personaTemplates/${id}`)
}

/** GET /api/personas/personaCategories */
export function getPersonaCategories() {
  return http.get<PersonaCategory[]>('/personas/personaCategories')
}
