import { useCallback, useEffect, useState } from 'react'
import { AlertCircle, LoaderCircle, ShieldCheck, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { LoadingState } from '@/components/shared/page'
import { useToast } from '@/components/feedback/toast-provider'
import { ApiError } from '@/lib/api'
import { permissionsApi, rolesApi, usersApi } from '@/lib/api-admin'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import type { Permission, Role, User, UserAccess } from '@/lib/api-types'

export function UserAccessDialog({ user, open, onOpenChange }: { user: User | null; open: boolean; onOpenChange: (open: boolean) => void }) {
  const { can } = useAuth()
  const { notify } = useToast()
  const [access, setAccess] = useState<UserAccess | null>(null)
  const [roles, setRoles] = useState<Role[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedRoles, setSelectedRoles] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [savingRoles, setSavingRoles] = useState(false)
  const [overridePermission, setOverridePermission] = useState('')
  const [overrideEffect, setOverrideEffect] = useState<'allow' | 'deny'>('allow')
  const [overrideNote, setOverrideNote] = useState('')
  const [overrideError, setOverrideError] = useState<string | null>(null)
  const [savingOverride, setSavingOverride] = useState(false)

  const canViewAccess = can(PERMISSIONS.adminUsersView)
  const canListRoles = can(PERMISSIONS.adminRolesManage)
  const canAssignRoles = can(PERMISSIONS.adminUsersAssignRoles)
  const canEditRoles = canListRoles && canAssignRoles
  const canManagePermissions = can(PERMISSIONS.adminPermissionsManage)

  const load = useCallback(async () => {
    if (!user || !canViewAccess) {
      setLoading(false)
      return
    }
    setLoading(true)
    setError(null)
    try {
      const [accessData, roleList, permissionList] = await Promise.all([
        usersApi.access(user.id),
        canListRoles ? rolesApi.list() : Promise.resolve([]),
        canManagePermissions ? permissionsApi.list() : Promise.resolve([]),
      ])
      setAccess(accessData)
      setRoles(roleList)
      setPermissions(permissionList)
      setSelectedRoles(roleList.filter(role => accessData.roles.includes(role.name)).map(role => role.id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load access details.')
    } finally {
      setLoading(false)
    }
  }, [canListRoles, canManagePermissions, canViewAccess, user])

  useEffect(() => { if (open) void load() }, [load, open])

  async function saveRoles() {
    if (!user) return
    setSavingRoles(true)
    setError(null)
    try {
      await usersApi.assignRoles(user.id, selectedRoles)
      notify('Roles updated.')
      await load()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Unable to update roles.')
    } finally {
      setSavingRoles(false)
    }
  }

  async function addOverride() {
    if (!user || !overridePermission) return
    setSavingOverride(true)
    setOverrideError(null)
    try {
      await usersApi.setOverride(user.id, { permission_id: overridePermission, effect: overrideEffect, note: overrideNote.trim() || undefined })
      setOverridePermission(''); setOverrideNote('')
      notify('Permission override saved.')
      await load()
    } catch (err) {
      setOverrideError(err instanceof ApiError ? err.message : 'Unable to save the override.')
    } finally {
      setSavingOverride(false)
    }
  }

  async function removeOverride(permissionId: string) {
    if (!user) return
    try {
      await usersApi.removeOverride(user.id, permissionId)
      notify('Permission override removed.')
      await load()
    } catch (err) {
      setOverrideError(err instanceof ApiError ? err.message : 'Unable to remove the override.')
    }
  }

  return <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle>Roles & access</DialogTitle>
        <DialogDescription>{user ? `${user.first_name} ${user.last_name} · ${user.email}` : ''}</DialogDescription>
      </DialogHeader>
      {loading ? <LoadingState label="Loading access…" /> : error && !access ? <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div> : access && <div className="space-y-5">
        {error && <div className="flex items-start gap-2 rounded-xl border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive"><AlertCircle size={14} className="shrink-0" aria-hidden="true" />{error}</div>}
        {(roles.length > 0 || canEditRoles) && <section>
          <h3 className="text-sm font-semibold">Assigned roles</h3>
          <p className="mt-0.5 text-[11px] text-muted-foreground">Roles grant the baseline permissions for this account.</p>
          <div className="mt-3 grid gap-2 sm:grid-cols-2">
            {roles.map(role => <label key={role.id} className="flex items-start gap-2 rounded-xl border border-border px-3 py-2 text-xs">
              <input type="checkbox" className="mt-0.5 size-3.5 accent-neutral-950" checked={selectedRoles.includes(role.id)} disabled={!canEditRoles} onChange={event => setSelectedRoles(current => event.target.checked ? [...current, role.id] : current.filter(id => id !== role.id))} />
              <span><span className="block font-semibold capitalize">{role.name}</span><span className="block text-[10px] text-muted-foreground">{role.scope === 'doctor' ? 'Doctor scope' : 'Admin scope'}{role.description ? ` · ${role.description}` : ` · ${role.permissions.length} permissions`}</span></span>
            </label>)}
          </div>
          {canEditRoles && <Button size="sm" className="mt-3" onClick={() => void saveRoles()} disabled={savingRoles}>{savingRoles && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Save roles</Button>}
        </section>}

        <section>
          <h3 className="flex items-center gap-1.5 text-sm font-semibold"><ShieldCheck size={14} aria-hidden="true" /> Effective permissions</h3>
          <div className="mt-2 flex flex-wrap gap-1">
            {access.permissions.map(permission => <span key={permission} className="rounded-full bg-muted px-2 py-0.5 text-[9px] font-medium text-muted-foreground">{permission}</span>)}
          </div>
        </section>


        <section>
          <h3 className="text-sm font-semibold">Overrides</h3>
          <p className="mt-0.5 text-[11px] text-muted-foreground">A deny override wins over role grants and allow overrides.</p>
          <div className="mt-2 space-y-1.5">
            {access.overrides.length === 0 ? <p className="text-[11px] text-muted-foreground">No explicit overrides.</p> : access.overrides.map(override => <div key={override.id} className="flex items-center justify-between gap-2 rounded-lg bg-muted/50 px-3 py-1.5 text-[11px]">
              <span className="min-w-0"><span className="font-medium">{override.permission}</span>{override.note && <span className="text-muted-foreground"> · {override.note}</span>}</span>
              <span className={`rounded-full px-1.5 py-0.5 text-[9px] font-semibold ${override.effect === 'deny' ? 'bg-destructive/10 text-destructive' : 'bg-emerald-50 text-emerald-700'}`}>{override.effect}</span>
              {canManagePermissions && <button type="button" aria-label="Remove override" className="text-muted-foreground hover:text-destructive" onClick={() => void removeOverride(override.permission_id)}><X size={13} aria-hidden="true" /></button>}
            </div>)}
          </div>
          {canManagePermissions && permissions.length > 0 && <div className="mt-3 grid gap-2 rounded-xl border border-border p-3 sm:grid-cols-[minmax(0,1fr)_100px]">
            <select value={overridePermission} onChange={event => setOverridePermission(event.target.value)} className="h-9 rounded-md border border-border bg-background px-2 text-[11px]">
              <option value="">Choose a permission…</option>
              {permissions.map(permission => <option key={permission.id} value={permission.id}>{permission.code}</option>)}
            </select>
            <select value={overrideEffect} onChange={event => setOverrideEffect(event.target.value as 'allow' | 'deny')} className="h-9 rounded-md border border-border bg-background px-2 text-[11px]">
              <option value="allow">Allow</option>
              <option value="deny">Deny</option>
            </select>
            <input value={overrideNote} onChange={event => setOverrideNote(event.target.value)} placeholder="Note (optional)" className="h-9 rounded-md border border-border bg-background px-2 text-[11px] sm:col-span-2" />
            {overrideError && <p className="text-[10px] text-destructive sm:col-span-2">{overrideError}</p>}
            <Button size="sm" className="sm:col-span-2" onClick={() => void addOverride()} disabled={!overridePermission || savingOverride}>{savingOverride && <LoaderCircle size={13} className="animate-spin" aria-hidden="true" />} Save override</Button>
          </div>}
        </section>
      </div>}
      <DialogFooter><Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button></DialogFooter>
    </DialogContent>
  </Dialog>
}

