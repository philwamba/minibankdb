'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function ReportPage() {
    type ReportRow = [number, string, number]
    const [rows, setRows] = useState<ReportRow[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const fetchReport = async () => {
        setLoading(true)
        try {
            const res = await fetch(`${API_URL}/api/reports/user-wallets`)
            if (!res.ok) throw new Error(`Request failed: ${res.status}`)
            
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            setRows(data.rows || [])
        } catch (err: any) {
            console.error(err)
            setError(err?.message || 'Failed to fetch report')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchReport()
    }, [])

    return (
        <div className="space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">User Wallets Report</h1>
                    <p className="text-slate-500 mt-1">Cross-reference users and their wallet balances (JOIN)</p>
                </div>
                <Button onClick={fetchReport} variant="outline">
                    <RefreshCw size={16} className="mr-2" />
                    Refresh
                </Button>
            </header>

            <Card>
                <CardContent className="p-0">
                    <table className="w-full text-sm text-left">
                        <thead className="bg-slate-50 text-slate-500 font-medium">
                            <tr>
                                <th className="px-4 py-3 border-b">User ID</th>
                                <th className="px-4 py-3 border-b">Name</th>
                                <th className="px-4 py-3 border-b">Balance</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr><td colSpan={3} className="p-4 text-center">Loading...</td></tr>
                            ) : rows.length === 0 ? (
                                <tr><td colSpan={3} className="p-4 text-center text-slate-400">No data found</td></tr>
                            ) : (
                                rows.map((row) => (
                                    <tr key={row[0]} className="border-b hover:bg-slate-50">
                                        <td className="px-4 py-3">{row[0]}</td>
                                        <td className="px-4 py-3 font-medium">{row[1]}</td>
                                        <td className="px-4 py-3 font-mono text-green-600">{row[2]}</td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </CardContent>
            </Card>
        </div>
    )
}
