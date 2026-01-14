'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Database, Terminal, Users, Wallet, ArrowRightLeft, BarChart3, Server, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'

export function Sidebar() {
    const pathname = usePathname()

    return (
        <aside className="w-64 bg-white border-r border-slate-200 hidden md:block">
            <div className="p-6">
                <div className="flex items-center gap-2 mb-8">
                    <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
                        <Database className="h-4 w-4 text-white" />
                    </div>
                    <span className="font-bold text-xl text-slate-900">
                        MiniBankDB
                    </span>
                </div>

                <nav className="space-y-1">
                    <NavItem
                        href="/"
                        icon={<Terminal size={20} />}
                        label="SQL Console"
                        active={pathname === '/'}
                    />
                    <div className="pt-4 pb-2">
                        <p className="px-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                            Application
                        </p>
                    </div>
                    <NavItem
                        href="/users"
                        icon={<Users size={20} />}
                        label="Users"
                        active={pathname === '/users'}
                    />
                    <NavItem
                        href="/wallets"
                        icon={<Wallet size={20} />}
                        label="Wallets"
                        active={pathname === '/wallets'}
                    />
                    <NavItem
                        href="/transactions"
                        icon={<ArrowRightLeft size={20} />}
                        label="Transactions"
                        active={pathname === '/transactions'}
                    />
                    <NavItem
                        href="/reports/user-wallets"
                        icon={<BarChart3 size={20} />}
                        label="Reports"
                        active={pathname === '/reports/user-wallets'}
                    />

                    <div className="pt-4 pb-2">
                        <p className="px-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                            System
                        </p>
                    </div>
                     <NavItem 
                        href="#" 
                        icon={<Server size={20} />} 
                        label="Storage" 
                    />
                    <NavItem
                        href="#"
                        icon={<Settings size={20} />}
                        label="Settings"
                    />
                </nav>
            </div>
        </aside>
    )
}

function NavItem({
    href,
    icon,
    label,
    active = false,
}: {
    href: string
    icon: React.ReactNode
    label: string
    active?: boolean
}) {
    return (
        <Link
            href={href}
            className={cn(
                'w-full flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                active
                    ? 'bg-blue-50 text-blue-700'
                    : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900',
            )}
        >
            {icon}
            {label}
        </Link>
    )
}
