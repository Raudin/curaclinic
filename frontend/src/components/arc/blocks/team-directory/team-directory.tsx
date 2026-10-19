import { useMemo, useState, type ReactNode } from 'react'
import { AnimatePresence, LayoutGroup, motion, useReducedMotion } from 'motion/react'
import { Check, Copy, Search } from 'lucide-react'
import { Avatar } from '@/components/arc/avatar/avatar'
import { Button } from '@/components/ui/button'
import styles from './team-directory.module.css'

export type TeamDirectoryPerson = {
  id: string
  name: string
  role: string
  team: string
  email: string
  timezone?: string
  focus?: string
  bio?: string
  avatarUrl?: string
  available?: boolean
}

export function TeamDirectory({
  people,
  onPersonSelect,
  renderActions,
}: {
  people: TeamDirectoryPerson[]
  onPersonSelect?: (person: TeamDirectoryPerson) => void
  renderActions?: (person: TeamDirectoryPerson) => ReactNode
}) {
  const reduceMotion = useReducedMotion() ?? false
  const [query, setQuery] = useState('')
  const [team, setTeam] = useState('All')
  const [selectedId, setSelectedId] = useState(people[0]?.id ?? '')
  const [copyMessage, setCopyMessage] = useState('')
  const teams = useMemo(() => [...new Set(people.map(person => person.team).filter(Boolean))].sort((a, b) => a.localeCompare(b)), [people])
  const counts = useMemo(() => new Map(teams.map(item => [item, people.filter(person => person.team === item).length])), [people, teams])
  const filtered = useMemo(() => {
    const search = query.trim().toLocaleLowerCase()
    return people.filter(person => {
      const matchesTeam = team === 'All' || person.team === team
      const searchable = [person.name, person.role, person.team, person.email, person.timezone, person.focus, person.bio].filter(Boolean).join(' ').toLocaleLowerCase()
      return matchesTeam && (!search || searchable.includes(search))
    })
  }, [people, query, team])
  const selected = filtered.find(person => person.id === selectedId) ?? filtered[0]

  async function copyEmail(email: string) {
    try {
      await navigator.clipboard.writeText(email)
      setCopyMessage('Email copied')
    } catch {
      setCopyMessage('Unable to copy email')
    }
    window.setTimeout(() => setCopyMessage(''), 2000)
  }

  function selectPerson(person: TeamDirectoryPerson) {
    setSelectedId(person.id)
    onPersonSelect?.(person)
  }

  return <section className={styles.directory} aria-label="Doctor directory">
    <div className={styles.header}>
      <h2 className={styles.heading}>Doctors <span>{people.length}</span></h2>
      <label className={styles.search}>
        <Search size={15} aria-hidden="true" />
        <input
          type="search"
          value={query}
          onChange={event => setQuery(event.target.value)}
          onKeyDown={event => { if (event.key === 'Escape') setQuery('') }}
          placeholder="Search name, specialty, email"
          aria-label="Search doctors by name, specialty, or email"
        />
      </label>
    </div>

    <div className={styles.filters} role="group" aria-label="Filter doctors by specialty">
      {[{ name: 'All', count: people.length }, ...teams.map(name => ({ name, count: counts.get(name) ?? 0 }))].map(item => <button
        key={item.name}
        type="button"
        className={styles.filter}
        aria-pressed={team === item.name}
        onClick={() => setTeam(item.name)}
      >
        {team === item.name && <motion.span className={styles.filterSelection} layoutId="team-directory-filter" transition={reduceMotion ? { duration: 0 } : { type: 'spring', stiffness: 420, damping: 36 }} />}
        <span className={styles.filterLabel}>{item.name}</span><span className={styles.count}>{item.count}</span>
      </button>)}
    </div>

    <div className={styles.body}>
      <div className={styles.people}>
        <LayoutGroup id="team-directory-people">
          <AnimatePresence initial={false} mode="popLayout">
            {filtered.map(person => <motion.button
              key={person.id}
              type="button"
              className={styles.person}
              aria-pressed={selected?.id === person.id}
              onClick={() => selectPerson(person)}
              initial={reduceMotion ? { opacity: 0 } : { opacity: 0, y: 5 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: reduceMotion ? 0 : -3 }}
              layout={!reduceMotion}
              transition={reduceMotion ? { duration: 0 } : { duration: .18 }}
            >
              {selected?.id === person.id && <motion.span className={styles.personSelection} layoutId="team-directory-person" transition={reduceMotion ? { duration: 0 } : { type: 'spring', stiffness: 420, damping: 36 }} />}
              <span className={styles.personAvatar}>
                <Avatar name={person.name} src={person.avatarUrl} size="md" />
                <span className={`${styles.presence} ${person.available ? styles.available : ''}`} aria-label={person.available ? 'Accepting patients' : 'Not accepting patients'} />
              </span>
              <span className={styles.personCopy}><span className={styles.personName}>{person.name}</span><span className={styles.personRole}>{person.role}</span></span>
            </motion.button>)}
          </AnimatePresence>
        </LayoutGroup>
        {filtered.length === 0 && <p className={styles.empty}>No doctors match your search.</p>}
      </div>

      <div className={styles.profile} aria-live="polite" aria-atomic="true">
        <AnimatePresence mode="wait" initial={false}>
          {selected ? <motion.div
            key={selected.id}
            className={styles.profileInner}
            initial={reduceMotion ? { opacity: 0 } : { opacity: 0, y: 4 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: reduceMotion ? 0 : -3 }}
            transition={reduceMotion ? { duration: 0 } : { duration: .18 }}
          >
            <div className={styles.profileHeading}>
              <Avatar name={selected.name} src={selected.avatarUrl} size="lg" />
              <div className={styles.profileTitle}><h3>{selected.name}</h3><p>{selected.role}</p></div>
            </div>
            <p className={styles.bio}>{selected.bio || `${selected.name} is part of the clinic team. Contact the clinic to learn more.`}</p>
            <dl className={styles.details}>
              <div><dt>Status</dt><dd><span className={`${styles.statusDot} ${selected.available ? styles.available : ''}`} />{selected.available ? 'Accepting patients' : 'Not accepting patients'}</dd></div>
              <div><dt>Specialty</dt><dd>{selected.team}</dd></div>
              {selected.focus && <div><dt>Practice</dt><dd>{selected.focus}</dd></div>}
              {selected.timezone && <div><dt>Timezone</dt><dd>{selected.timezone.replaceAll('_', ' ')}</dd></div>}
              <div className={styles.emailRow}><dt>Email</dt><dd><a href={`mailto:${selected.email}`}>{selected.email}</a><Button type="button" variant="ghost" size="sm" className={styles.copy} aria-label={`Copy ${selected.email}`} onClick={() => void copyEmail(selected.email)}>{copyMessage === 'Email copied' ? <Check size={15} aria-hidden="true" /> : <Copy size={15} aria-hidden="true" />}</Button></dd></div>
            </dl>
            <p className={styles.announcement} role="status">{copyMessage}</p>
            {renderActions?.(selected) && <div className={styles.actions}>{renderActions(selected)}</div>}
          </motion.div> : <p className={styles.empty}>Select a doctor to view their profile.</p>}
        </AnimatePresence>
      </div>
    </div>
  </section>
}
