import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Sidebar } from '@/components/Sidebar'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
    title: 'MiniBankDB Console',
    description: 'Database Engine Demo',
}

export default function RootLayout({
    children,
}: Readonly<{
    children: React.ReactNode
}>) {
    return (
        <html lang="en">
            <body className={inter.className}>
                <div className="min-h-screen bg-slate-50 flex">
                    <Sidebar />
                    <main className="flex-1 p-8 overflow-auto">
                        {children}
                    </main>
                </div>
            </body>
        </html>
    )
}
