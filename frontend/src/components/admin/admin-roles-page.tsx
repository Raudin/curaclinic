import { useCallback, useEffect, useMemo, useState } from 'react'
import { Copy, LoaderCircle, Pencil, Plus, ShieldCheck, Trash2 } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { ErrorState, LoadingState, PageHeader } from '@/components/shared/page'
import { RolesPermissions } from '@/components/arc/blocks/roles-permissions/roles-permissions'
import { useToast } from '@/components/feedback/toast-provider'
import { ApiError } from '@/lib/api'
import { permissionsApi, rolesApi } from '@/lib/api-admin'
import { permissionScope } from '@/lib/permissions'
import type { Permission, Role, RoleScope } from '@/lib/api-types'

function samePermissions(left: string[], right: string[]) {
  return left.length === right.length && left.every(permission => right.includes(permission))
}

export function AdminRolesPage() {
  const { notify } = useToast()
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [drafts, setDrafts] = useState<Record<string, string[]>>({})
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [editing, setEditing] = useState<Role | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [dialogOpen, setDialogOpen] = useState(false)
  const [scope, setScope] = useState<RoleScope>('doctor')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [saving, setSaving] = useState(false)
  const [dialogError, setDialogError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [roleList, permissionList] = await Promise.all([rolesApi.list(), permissionsApi.list()])
      setRoles(roleList)
      setPermissions(permissionList)
      setDrafts(Object.fromEntries(roleList.map(role => [role.id, role.name === 'admin'
        ? permissionList.filter(permission => permission.scope === role.scope).map(permission => permission.code)
        : [...role.permissions]])))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load roles.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { void load() }, [load])

  const pendingRoles = useMemo(() => roles.filter(role => role.name !== 'admin' && !samePermissions(drafts[role.id] ?? role.permissions, role.permissions)), [drafts, roles])
  const dialogScope = editing?.scope ?? scope
  const dialogPermissions = useMemo(() => permissions.filter(permission => permission.scope === dialogScope), [dialogScope, permissions])
  const grouped = useMemo(() => dialogPermissions.reduce<Record<string, Permission[]>>((groups, permission) => {
    groups[permission.resource] = [...(groups[permission.resource] ?? []), permission]
    return groups
  }, {}), [dialogPermissions])

  function openEdit(role: Role) {
    setEditing(role)
    setScope(role.scope)
    setName(role.name)
    setDescription(role.description ?? '')
    setSelected(role.name === 'admin'
      ? permissions.filter(permission => permission.scope === role.scope).map(permission => permission.code)
      : [...(drafts[role.id] ?? role.permissions)])
    setDialogError(null)
    setDialogOpen(true)
  }

  function openCreate() {
    setEditing(null)
    setScope('doctor')
    setName('')
    setDescription('')
    setSelected([])
    setDialogError(null)
    setDialogOpen(true)
  }

  function duplicateRole(role: Role) {
    setEditing(null)
    setScope(role.scope)
    setName(`${role.name} copy`)
    setDescription(role.description ? `Copy of ${role.description}` : '')
    setSelected((drafts[role.id] ?? role.permissions).filter(code => permissionScope(code) === role.scope))
    setDialogError(null)
    setDialogOpen(true)
  }

  async function saveRole() {
    setSaving(true)
    setDialogError(null)
    try {
      const permissionIds = permissions
        .filter(permission => permission.scope === dialogScope && selected.includes(permission.code))
        .map(permission => permission.id)
      if (editing) {
        const updated = await rolesApi.update(editing.id, { name: editing.is_system ? editing.name : name.trim(), description: description.trim(), permission_ids: permissionIds })
        setRoles(current => current.map(role => role.id === updated.id ? updated : role))
        setDrafts(current => ({ ...current, [updated.id]: [...updated.permissions] }))
        notify('Role updated.')
      } else {
        await rolesApi.create({ scope, name: name.trim(), description: description.trim(), permission_ids: permissionIds })
        notify('Role created.')
        await load()
      }
      setDialogOpen(false)
    } catch (err) {
      setDialogError(err instanceof ApiError ? err.message : 'Unable to save the role.')
    } finally {
      setSaving(false)
    }
  }

  function togglePermission(role: Role, code: string) {
    if (role.name === 'admin' || permissionScope(code) !== role.scope) return
    setDrafts(current => {
      const currentCodes = current[role.id] ?? role.permissions
      return { ...current, [role.id]: currentCodes.includes(code) ? currentCodes.filter(item => item !== code) : [...currentCodes, code] }
    })
    setSaveError(null)
  }

  function toggleGroup(role: Role, codes: string[], checked: boolean) {
    if (role.name === 'admin') return
    const scopedCodes = codes.filter(code => permissionScope(code) === role.scope)
    setDrafts(current => {
      const next = new Set(current[role.id] ?? role.permissions)
      scopedCodes.forEach(code => checked ? next.add(code) : next.delete(code))
      return { ...current, [role.id]: [...next] }
    })
    setSaveError(null)
  }

  function discardChanges() {
    setDrafts(Object.fromEntries(roles.map(role => [role.id, role.name === 'admin'
      ? permissions.filter(permission => permission.scope === role.scope).map(permission => permission.code)
      : [...role.permissions]])))
    setSaveError(null)
  }

  async function saveChanges() {
    if (pendingRoles.length === 0) return
    setSaving(true)
    setSaveError(null)
    const updatedRoles: Role[] = []
    const failedRoles: string[] = []
    let firstError = ''
    for (const role of pendingRoles) {
      const codes = drafts[role.id] ?? role.permissions
      const permissionIds = permissions
        .filter(permission => permission.scope === role.scope && codes.includes(permission.code))
        .map(permission => permission.id)
      try {
        updatedRoles.push(await rolesApi.update(role.id, { name: role.name, description: role.description ?? '', permission_ids: permissionIds }))
      } catch (err) {
        failedRoles.push(role.name)
        if (!firstError) firstError = err instanceof ApiError ? err.message : 'Unable to save role changes.'
      }
    }
    if (updatedRoles.length) {
      setRoles(current => current.map(role => updatedRoles.find(updated => updated.id === role.id) ?? role))
      setDrafts(current => ({ ...current, ...Object.fromEntries(updatedRoles.map(role => [role.id, [...role.permissions]])) }))
    }
    if (failedRoles.length) {
      setSaveError(`${firstError} Changes could not be saved for: ${failedRoles.join(', ')}.`)
      if (updatedRoles.length) notify(`${updatedRoles.length} role${updatedRoles.length === 1 ? '' : 's'} saved. Some changes need attention.`, 'Partially saved')
    } else {
      notify('Role permissions saved.')
    }
    setSaving(false)
  }

  async function remove(role: Role) {
    if (role.is_system || !window.confirm(`Delete the ${role.name} role?`)) return
    try {
      await rolesApi.remove(role.id)
      notify('Role deleted.')
      await load()
    } catch (err) {
      notify(err instanceof ApiError ? err.message : 'Unable to delete the role.', 'Delete failed')
    }
  }

  return <main className="mx-auto w-full max-w-[1200px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader eyebrow="Access control" title="Roles & permissions" description="Control which parts of CuraClinic each role can access." actions={<Button className="gap-2" onClick={openCreate} disabled={saving || pendingRoles.length > 0}><Plus size={15} aria-hidden="true" /> New role</Button>} />
    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading roles…" /> : null}
    {!loading && !error && <>
      <div className="mb-3 flex items-center gap-2 text-xs text-muted-foreground"><ShieldCheck size={14} aria-hidden="true" /> CuraClinic · {roles.length} roles</div>
      {saveError && <p role="alert" className="mb-3 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive">{saveError}</p>}
      <RolesPermissions
        roles={roles}
        permissions={permissions}
        grants={drafts}
        disabled={saving}
        onTogglePermission={togglePermission}
        onToggleGroup={toggleGroup}
        renderRoleActions={role => <>
          <Button type="button" variant="ghost" size="sm" className="h-7 w-7 p-0" aria-label={`Edit ${role.name} role`} title="Edit role" onClick={() => openEdit(role)} disabled={saving || pendingRoles.length > 0}><Pencil size={12} aria-hidden="true" /></Button>
          <Button type="button" variant="ghost" size="sm" className="h-7 w-7 p-0" aria-label={`Duplicate ${role.name} role`} title="Duplicate role" onClick={() => duplicateRole(role)} disabled={saving || pendingRoles.length > 0}><Copy size={12} aria-hidden="true" /></Button>
          {!role.is_system && <Button type="button" variant="ghost" size="sm" className="h-7 w-7 p-0 text-muted-foreground hover:text-destructive" aria-label={`Delete ${role.name} role`} title="Delete role" onClick={() => void remove(role)} disabled={saving || pendingRoles.length > 0}><Trash2 size={12} aria-hidden="true" /></Button>}
        </>}
      />
      </>}

    <AnimateReviewBar dirty={pendingRoles.length > 0} saving={saving} count={pendingRoles.length} onDiscard={discardChanges} onSave={() => void saveChanges()} />

    <Dialog open={dialogOpen} onOpenChange={open => { if (!open && !saving) setDialogOpen(false) }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader><DialogTitle>{editing ? `Edit ${editing.name}` : 'Create a role'}</DialogTitle><DialogDescription>{editing ? 'Update the role details and its permission grants.' : 'Create a role with the permissions your team needs.'}</DialogDescription></DialogHeader>
        {dialogError && <p role="alert" className="rounded-lg bg-destructive/5 px-3 py-2 text-xs text-destructive">{dialogError}</p>}
        <div className="grid gap-2 sm:grid-cols-3">
          <input aria-label="Role name" value={name} onChange={event => setName(event.target.value)} placeholder="Role name" disabled={saving || Boolean(editing?.is_system)} className="h-9 rounded-md border border-border bg-background px-2 text-xs disabled:opacity-60" />
          <input aria-label="Role description" value={description} onChange={event => setDescription(event.target.value)} placeholder="Description" disabled={saving} className="h-9 rounded-md border border-border bg-background px-2 text-xs disabled:opacity-60" />
          <select
            aria-label="Role scope"
            value={editing?.scope ?? scope}
            onChange={event => { setScope(event.target.value as RoleScope); setSelected([]) }}
            disabled={saving || Boolean(editing)}
            className="h-9 rounded-md border border-border bg-background px-2 text-xs disabled:opacity-60"
          >
            <option value="doctor">Doctor scope</option>
            <option value="admin">Admin scope</option>
          </select>
        </div>
        <div className="max-h-[45vh] space-y-3 overflow-y-auto rounded-xl border border-border p-3">
          {Object.entries(grouped).map(([resource, list]) => <fieldset key={resource} disabled={saving || editing?.name === 'admin'}>
            <legend className="text-[9px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">{resource}</legend>
            <div className="mt-1 grid gap-1 sm:grid-cols-2">{list.map(permission => <label key={permission.id} className="flex items-center gap-2 text-[11px]"><input type="checkbox" className="size-3.5 accent-neutral-950" checked={selected.includes(permission.code)} disabled={saving} onChange={event => setSelected(current => event.target.checked ? [...current, permission.code] : current.filter(code => code !== permission.code))} /><span>{permission.action.replaceAll('_', ' ')}</span></label>)}</div>
          </fieldset>)}
        </div>
        <DialogFooter><Button variant="outline" onClick={() => setDialogOpen(false)} disabled={saving}>Cancel</Button><Button onClick={() => void saveRole()} disabled={saving || !name.trim()}>{saving && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} {editing ? 'Save role' : 'Create role'}</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  </main>
}

function AnimateReviewBar({ dirty, saving, count, onDiscard, onSave }: { dirty: boolean; saving: boolean; count: number; onDiscard: () => void; onSave: () => void }) {
  const reduced = useReducedMotion() ?? false
  return <AnimatePresence>{dirty && <motion.div className="fixed inset-x-0 bottom-4 z-40 mx-auto flex w-[min(640px,calc(100%-2rem))] items-center justify-between gap-4 rounded-2xl border border-border bg-card px-4 py-3 shadow-xl sm:bottom-6 sm:px-5" initial={{ opacity: 0, y: reduced ? 0 : 18 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: reduced ? 0 : 18 }} transition={{ duration: reduced ? 0 : .18 }}>
    <p className="text-xs font-medium">Unsaved changes <span className="text-muted-foreground">· {count} role{count === 1 ? '' : 's'} updated</span></p>
    <div className="flex shrink-0 gap-2"><Button variant="ghost" size="sm" onClick={onDiscard} disabled={saving}>Discard</Button><Button size="sm" className="gap-1.5" onClick={onSave} disabled={saving}>{saving && <LoaderCircle size={12} className="animate-spin" aria-hidden="true" />} Save changes</Button></div>
  </motion.div>}</AnimatePresence>
}
