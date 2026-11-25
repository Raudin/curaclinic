import { useCallback, useEffect, useState } from 'react'
import { KeyRound, Search, ShieldCheck, UserCheck, UserX } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Button } from '@/components/ui/button'
import { EmptyState, ErrorState, LoadingState, PageHeader, Pagination } from '@/components/shared/page'
import { UserAccessDialog } from '@/components/admin/user-access-dialog'
import { useToast } from '@/components/feedback/toast-provider'
import { useAuth } from '@/lib/auth-context'
import { PERMISSIONS } from '@/lib/permissions'
import { usersApi } from '@/lib/api-admin'
import { ApiError } from '@/lib/api'
import type { User } from '@/lib/api-types'
import { formatDate } from '@/lib/time'

const PAGE_SIZE = 20

export function AdminPatientsPage() {
  const { can } = useAuth()
  const { notify } = useToast()
  const [search, setSearch] = useState('')
  const [query, setQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState('patient')
  const [items, setItems] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [offset, setOffset] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [accessUser, setAccessUser] = useState<User | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const page = await usersApi.list({ role: roleFilter || undefined, search: query || undefined, limit: PAGE_SIZE, offset })
      setItems(page.data)
      setTotal(page.pagination.total)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to load patients.')
    } finally {
      setLoading(false)
    }
  }, [offset, query, roleFilter])

  useEffect(() => { void load() }, [load])

  async function toggleStatus(user: User) {
    try {
      await usersApi.setStatus(user.id, user.status === 'active' ? 'inactive' : 'active')
      notify(`${user.first_name} ${user.last_name} is now ${user.status === 'active' ? 'inactive' : 'active'}.`)
      await load()
    } catch (err) {
      notify(err instanceof ApiError ? err.message : 'Unable to update this account.', 'Update failed')
    }
  }

  return <main className="mx-auto w-full max-w-[1300px] px-4 pb-12 pt-7 sm:px-7 sm:pt-9 lg:px-10">
    <PageHeader title="Patients" description="Search patient accounts, manage their status and review role-based access." />
    <form className="mb-4 flex flex-wrap gap-2" onSubmit={event => { event.preventDefault(); setOffset(0); setQuery(search.trim()) }}>
      <div className="relative min-w-[240px] flex-1">
        <Search size={14} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" aria-hidden="true" />
        <input value={search} onChange={event => setSearch(event.target.value)} placeholder="Search by name or email" className="h-10 w-full rounded-xl border border-border bg-card pl-9 pr-3 text-xs outline-none focus:ring-2 focus:ring-ring" aria-label="Search patients" />
      </div>
      <select value={roleFilter} onChange={event => { setRoleFilter(event.target.value); setOffset(0) }} className="h-10 rounded-xl border border-border bg-card px-3 text-xs" aria-label="Filter by role">
        <option value="patient">Patients</option>
        <option value="doctor">Doctors</option>
        <option value="admin">Admins</option>
        <option value="">All roles</option>
      </select>
      <Button type="submit" variant="outline">Search</Button>
      {query && <Button type="button" variant="ghost" onClick={() => { setSearch(''); setQuery(''); setOffset(0) }}>Clear</Button>}
    </form>

    {error && <ErrorState message={error} onRetry={load} />}
    {loading && !error ? <LoadingState label="Loading patients…" /> : null}
    {!loading && !error && items.length === 0 ? <EmptyState title={query ? "No matching accounts" : "No accounts found"} description={query ? "Try another name or email, or clear the search." : "Accounts matching the selected role will appear here."} /> : null}
    {!loading && !error && items.length > 0 && <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="@container/directory">
        <table className="block w-full text-left text-sm @min-[760px]/directory:table">
          <thead className="hidden @min-[760px]/directory:table-header-group border-b border-border bg-muted/40 text-xs uppercase tracking-[0.08em] text-muted-foreground">
            <tr><th className="px-4 py-3 font-semibold">Patient</th><th className="px-4 py-3 font-semibold">Roles</th><th className="px-4 py-3 font-semibold">Status</th><th className="px-4 py-3 font-semibold">Joined</th><th className="px-4 py-3 text-right font-semibold">Actions</th></tr>
          </thead>
          <tbody className="block @min-[760px]/directory:table-row-group divide-y divide-border/70">
            {items.map(user => <tr key={user.id} className="grid gap-2 px-4 py-4 align-middle @min-[760px]/directory:table-row @max-[760px]/directory:[&>td]:p-0 [&>td]:break-words">
              <td className="px-4 py-3"><div className="flex items-center gap-2.5"><Avatar name={`${user.first_name} ${user.last_name}`} size="sm" /><div><p className="font-semibold">{user.first_name} {user.last_name}</p><p className="text-xs text-muted-foreground">{user.email}</p></div></div></td>
              <td className="px-4 py-3"><div className="flex flex-wrap gap-1">{user.roles.map(role => <span key={role} className="rounded-full bg-muted px-2 py-0.5 text-xs font-medium capitalize text-muted-foreground">{role}</span>)}</div></td>
              <td className="px-4 py-3"><span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-semibold capitalize ${user.status === 'active' ? 'bg-emerald-50 text-emerald-700' : 'bg-muted text-muted-foreground'}`}>{user.status === 'active' ? <UserCheck size={10} aria-hidden="true" /> : <UserX size={10} aria-hidden="true" />}{user.status}</span></td>
              <td className="px-4 py-3 text-xs text-muted-foreground"><span className="@min-[760px]/directory:hidden">Joined </span>{formatDate(user.created_at, 'Africa/Nairobi', { weekday: undefined })}</td>
              <td className="px-4 py-3"><div className="flex flex-wrap gap-2 @min-[760px]/directory:justify-end">
                <Button variant="outline" size="sm" className="min-h-10 gap-1" onClick={() => setAccessUser(user)}><KeyRound size={12} aria-hidden="true" /> Access</Button>
                {can(PERMISSIONS.adminUsersUpdate) && <Button variant="ghost" size="sm" className="min-h-10 gap-1" onClick={() => void toggleStatus(user)}>{user.status === 'active' ? <UserX size={12} aria-hidden="true" /> : <ShieldCheck size={12} aria-hidden="true" />}{user.status === 'active' ? 'Deactivate' : 'Activate'}</Button>}
              </div></td>
            </tr>)}
          </tbody>
        </table>
      </div>
      {total > PAGE_SIZE && <div className="border-t border-border px-4 py-2"><Pagination limit={PAGE_SIZE} offset={offset} total={total} onChange={setOffset} /></div>}
    </div>}
    <UserAccessDialog user={accessUser} open={Boolean(accessUser)} onOpenChange={open => { if (!open) setAccessUser(null) }} />
  </main>
}

