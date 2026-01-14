'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Play, Terminal } from 'lucide-react'
import { motion } from 'framer-motion'

export default function Dashboard() {
    const [query, setQuery] = useState('SELECT * FROM accounts;')
    const [results, setResults] = useState<{
        columns: string[]
        rows: any[][]
    } | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [loading, setLoading] = useState(false)

    const runQuery = async () => {
        setLoading(true)
        setError(null)
        try {
            const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
            const res = await fetch(`${API_URL}/api/query`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ query }),
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            setResults(data)
        } catch (err: any) {
            setError(err.message)
            setResults(null)
        } finally {
            setLoading(false)
        }
    }

    return (
        <>
            <header className="mb-8 flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">
                        SQL Console
                    </h1>
                    <p className="text-slate-500 mt-1">
                        Execute queries against your MiniBankDB instance.
                    </p>
                </div>
            </header>

            <div className="grid gap-6">
                    {/* Query Editor */}
                    <Card className="border-slate-200 shadow-sm overflow-hidden">
                        <CardHeader className="bg-slate-50 border-b border-slate-100 py-4">
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-2 text-sm font-medium text-slate-700">
                                    <Terminal
                                        size={16}
                                        className="text-blue-600"
                                    />
                                    <span>Query Editor</span>
                                </div>
                                <div className="flex gap-2">
                                    <Button
                                        size="sm"
                                        variant="outline"
                                        onClick={() =>
                                            setQuery(
                                                "INSERT INTO accounts (id, name, balance) VALUES (1, 'Alice', 1000);",
                                            )
                                        }
                                    >
                                        Insert Sample
                                    </Button>
                                    <Button
                                        size="sm"
                                        onClick={runQuery}
                                        disabled={loading}
                                        className="bg-blue-600 hover:bg-blue-700"
                                    >
                                        <Play size={16} className="mr-2" />
                                        {loading ? 'Running...' : 'Run Query'}
                                    </Button>
                                </div>
                            </div>
                        </CardHeader>
                        <div className="p-0">
                            <textarea
                                value={query}
                                onChange={e => setQuery(e.target.value)}
                                className="w-full h-40 p-4 font-mono text-sm focus:outline-none resize-none bg-white text-slate-800"
                                placeholder="Enter your SQL query..."
                            />
                        </div>
                    </Card>

                    {/* Results Area */}
                    <motion.div
                        initial={{ opacity: 0, y: 10 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.3 }}
                    >
                        {error && (
                            <div className="bg-red-50 text-red-600 p-4 rounded-lg border border-red-100 flex items-center gap-2">
                                <span>Error:</span>
                                <span className="font-mono text-sm">
                                    {error}
                                </span>
                            </div>
                        )}

                        {results && (
                            <Card>
                                <CardHeader>
                                    <CardTitle className="text-lg">
                                        Query Results
                                    </CardTitle>
                                </CardHeader>
                                <CardContent>
                                    <div className="rounded-md border overflow-hidden">
                                        <table className="w-full text-sm text-left">
                                            <thead className="bg-slate-50 text-slate-500 font-medium">
                                                <tr>
                                                    {results.columns?.map(
                                                        (col, i) => (
                                                            <th
                                                                key={i}
                                                                className="px-4 py-3 border-b"
                                                            >
                                                                {col}
                                                            </th>
                                                        ),
                                                    )}
                                                </tr>
                                            </thead>
                                            <tbody>
                                                {results.rows?.length === 0 ? (
                                                    <tr>
                                                        <td
                                                            colSpan={
                                                                results.columns
                                                                    ?.length ||
                                                                1
                                                            }
                                                            className="px-4 py-8 text-center text-slate-400"
                                                        >
                                                            No rows returned
                                                        </td>
                                                    </tr>
                                                ) : (
                                                    results.rows?.map(
                                                        (row, i) => (
                                                            <tr
                                                                key={i}
                                                                className="border-b last:border-0 hover:bg-slate-50/50 transition-colors"
                                                            >
                                                                {row.map(
                                                                    (
                                                                        cell: any,
                                                                        j: number,
                                                                    ) => (
                                                                        <td
                                                                            key={
                                                                                j
                                                                            }
                                                                            className="px-4 py-3"
                                                                        >
                                                                            {
                                                                                cell
                                                                            }
                                                                        </td>
                                                                    ),
                                                                )}
                                                            </tr>
                                                        ),
                                                    )
                                                )}
                                            </tbody>
                                        </table>
                                    </div>
                                    <div className="mt-4 text-xs text-slate-400 text-right">
                                        {results.rows?.length || 0} rows found
                                    </div>
                                </CardContent>
                            </Card>
                        )}
                    </motion.div>
                </div>
        </>
    )
}
