import { useState, type ReactNode } from 'react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { ChevronDown, ChevronRight, LockKeyhole, Minus } from 'lucide-react'
import type { Permission, Role } from '@/lib/api-types'
import styles from './roles-permissions.module.css'

export function RolesPermissions({
  roles,
  permissions,
  grants,
  disabled = false,
  onTogglePermission,
  onToggleGroup,
  renderRoleActions,
}: {
  roles: Role[]
  permissions: Permission[]
  grants: Record<string, string[]>
  disabled?: boolean
  onTogglePermission: (role: Role, code: string) => void
  onToggleGroup: (role: Role, codes: string[], checked: boolean) => void
  renderRoleActions?: (role: Role) => ReactNode
}) {
  const reduced = useReducedMotion() ?? false
  const [collapsed, setCollapsed] = useState<Set<string>>(() => new Set())
  const grouped = permissions.reduce<Record<string, Permission[]>>((result, permission) => {
    result[permission.resource] = [...(result[permission.resource] ?? []), permission]
    return result
  }, {})
  const groups = Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b))

  return <div className={styles.shell}>
    <div className={styles.scroll}>
      <table className={styles.matrix}>
        <thead>
          <tr>
            <th scope="col" className={styles.permissionHeading}>Permission</th>
            {roles.map(role => {
              const roleCodes = permissions.filter(permission => permission.scope === role.scope).map(permission => permission.code)
              const selected = role.name === 'admin' ? roleCodes : (grants[role.id] ?? role.permissions).filter(code => roleCodes.includes(code))
              const permissionCount = selected.length
              return <th key={role.id} scope="col" className={styles.roleHeading}>
              <div className={styles.roleTitle}>
                <span className={styles.roleName}>{role.name}</span>
                {role.is_system && <span className={styles.system}>System</span>}
              </div>
              <span className={styles.roleCount}>{role.scope} scope · {permissionCount} {permissionCount === 1 ? 'permission' : 'permissions'}</span>
              {renderRoleActions?.(role) && <div className={styles.roleActions}>{renderRoleActions(role)}</div>}
              </th>
            })}
          </tr>
        </thead>
        <tbody>
          {groups.map(([resource, items]) => {
            const isCollapsed = collapsed.has(resource)
            return <PermissionGroup
              key={resource}
              resource={resource}
              items={items}
              roles={roles}
              grants={grants}
              collapsed={isCollapsed}
              reduced={reduced}
              onToggleCollapse={() => setCollapsed(current => {
                const next = new Set(current)
                if (next.has(resource)) next.delete(resource)
                else next.add(resource)
                return next
              })}
              onTogglePermission={onTogglePermission}
              onToggleGroup={onToggleGroup}
              disabled={disabled}
            />
          })}
        </tbody>
      </table>
      {permissions.length === 0 && <p className={styles.empty}>No permissions are available.</p>}
    </div>
    <div className={styles.footer}><LockKeyhole size={12} aria-hidden="true" /> Roles can only grant permissions from their own scope. The admin system role is locked.</div>
  </div>
}

function PermissionGroup({
  resource,
  items,
  roles,
  grants,
  collapsed,
  reduced,
  disabled,
  onToggleCollapse,
  onTogglePermission,
  onToggleGroup,
}: {
  resource: string
  items: Permission[]
  roles: Role[]
  grants: Record<string, string[]>
  collapsed: boolean
  reduced: boolean
  disabled: boolean
  onToggleCollapse: () => void
  onTogglePermission: (role: Role, code: string) => void
  onToggleGroup: (role: Role, codes: string[], checked: boolean) => void
}) {
  return <>
    <tr className={styles.groupRow}>
      <th scope="row" className={styles.groupLabel}>
        <button type="button" aria-expanded={!collapsed} onClick={onToggleCollapse} className={styles.groupButton}>
          {collapsed ? <ChevronRight size={14} aria-hidden="true" /> : <ChevronDown size={14} aria-hidden="true" />}
          <span>{resourceLabel(resource)}</span><span className={styles.groupCount}>{items.length}</span>
        </button>
      </th>
      {roles.map(role => {
        const scopedItems = items.filter(item => item.scope === role.scope)
        const roleCodes = scopedItems.map(item => item.code)
        const selected = role.name === 'admin' ? roleCodes : (grants[role.id] ?? role.permissions).filter(code => roleCodes.includes(code))
        const count = scopedItems.filter(item => selected.includes(item.code)).length
        const checked: boolean | 'mixed' = count === 0 ? false : count === scopedItems.length ? true : 'mixed'
        return <td key={role.id} className={styles.groupCell}>
          {scopedItems.length === 0
            ? <span className="text-muted-foreground" aria-hidden="true">—</span>
            : <PermissionToggle
              label={`${checked === true ? 'Remove' : 'Grant'} all ${resourceLabel(resource)} permissions for ${role.name}`}
              checked={checked}
              onClick={() => onToggleGroup(role, roleCodes, checked !== true)}
              disabled={disabled || role.name === 'admin'}
            />}
        </td>
      })}
    </tr>
    <AnimatePresence initial={false} mode="sync">
      {!collapsed && items.map((permission, index) => <motion.tr
        key={permission.id}
        className={styles.permissionRow}
        initial={reduced ? { opacity: 0 } : { opacity: 0, y: -3 }}
        animate={{ opacity: 1, y: 0 }}
        exit={reduced ? { opacity: 0, transition: { duration: 0 } } : { opacity: 0, y: -2, transition: { duration: .12 } }}
        transition={reduced ? { duration: 0 } : { duration: .16, delay: index * .018 }}
      >
        <th scope="row" className={styles.permissionName} title={permission.description || permission.code}>{permissionLabel(permission)}</th>
        {roles.map(role => {
          if (permission.scope !== role.scope) {
            return <td key={role.id} className={styles.permissionCell}><span className="text-muted-foreground" aria-hidden="true">—</span></td>
          }
          const selected = role.name === 'admin' ? [permission.code] : (grants[role.id] ?? role.permissions)
          return <td key={role.id} className={styles.permissionCell}>
            <PermissionToggle
              label={`${permissionLabel(permission)} for ${role.name}`}
              checked={selected.includes(permission.code)}
              onClick={() => onTogglePermission(role, permission.code)}
              disabled={disabled || role.name === 'admin'}
            />
          </td>
        })}
      </motion.tr>)}
    </AnimatePresence>
  </>
}

function PermissionToggle({ label, checked, onClick, disabled }: { label: string; checked: boolean | 'mixed'; onClick: () => void; disabled: boolean }) {
  return <button
    type="button"
    role="checkbox"
    aria-checked={checked}
    aria-label={label}
    disabled={disabled}
    onClick={onClick}
    className={styles.toggle}
    data-checked={checked}
  >
    {checked === 'mixed' ? <Minus size={12} strokeWidth={2.5} aria-hidden="true" /> : checked ? <span className={styles.check} aria-hidden="true" /> : null}
  </button>
}

function permissionLabel(permission: Permission) {
  if (permission.description) return permission.description
  const words = permission.action.replaceAll('_', ' ').replace(/\b\w/g, letter => letter.toUpperCase())
  const prefixes: Record<string, string> = {
    appointments: 'appointments',
    schedule: 'schedule',
    profiles: 'profiles',
    users: 'users',
    permissions: 'user permissions',
    doctors: 'doctor profiles',
    reports: 'reports',
    audit: 'audit logs',
  }
  const prefix = prefixes[permission.resource]
  return prefix ? `${words} ${prefix}` : `${words} ${resourceLabel(permission.resource).toLowerCase()}`
}

function resourceLabel(resource: string) {
  return resource.replaceAll('_', ' ').replace(/\b\w/g, letter => letter.toUpperCase())
}
