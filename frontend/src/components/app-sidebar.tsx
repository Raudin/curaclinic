import type { LucideIcon } from 'lucide-react'
import { HeartPulse } from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'

export type SidebarLink = {
  label: string
  path: string
  icon: LucideIcon
  matches?: (pathname: string) => boolean
}

type AppSidebarProps = {
  pathname: string
  onNavigate: (path: string) => void
  links: SidebarLink[]
  footerLink?: SidebarLink
  brand?: string
  homePath?: string
}

export function AppSidebar({ pathname, onNavigate, links, footerLink, brand = 'CuraClinic', homePath = '/' }: AppSidebarProps) {
  const { isMobile, setOpenMobile } = useSidebar()

  return <Sidebar collapsible="icon">
    <SidebarHeader><SidebarMenu><SidebarMenuItem><SidebarMenuButton asChild><a href={homePath} aria-label={`${brand} home`} onClick={(event) => { event.preventDefault(); onNavigate(homePath); if (isMobile) setOpenMobile(false) }}><span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground"><HeartPulse size={14} strokeWidth={2} aria-hidden="true" /></span><span className="text-sm font-semibold tracking-tight">{brand}</span></a></SidebarMenuButton></SidebarMenuItem></SidebarMenu></SidebarHeader>
    <SidebarContent><SidebarGroup><SidebarMenu>{links.map(({ label, path, icon: Icon, matches }) => {
      const active = matches ? matches(pathname) : pathname === path
      return <SidebarMenuItem key={path}><SidebarMenuButton asChild isActive={active} tooltip={label}><a href={path} aria-current={active ? 'page' : undefined} onClick={(event) => { event.preventDefault(); onNavigate(path); if (isMobile) setOpenMobile(false) }}><Icon /><span>{label}</span></a></SidebarMenuButton></SidebarMenuItem>
    })}</SidebarMenu></SidebarGroup></SidebarContent>
    {footerLink && <SidebarFooter><SidebarMenu><SidebarMenuItem><SidebarMenuButton asChild isActive={footerLink.matches ? footerLink.matches(pathname) : pathname === footerLink.path} tooltip={footerLink.label}><a href={footerLink.path} aria-current={(footerLink.matches ? footerLink.matches(pathname) : pathname === footerLink.path) ? 'page' : undefined} onClick={(event) => { event.preventDefault(); onNavigate(footerLink.path); if (isMobile) setOpenMobile(false) }}><footerLink.icon /><span>{footerLink.label}</span></a></SidebarMenuButton></SidebarMenuItem></SidebarMenu></SidebarFooter>}
    <SidebarRail />
  </Sidebar>
}
