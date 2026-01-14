'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, ArrowRightLeft } from 'lucide-react'

type Transaction = {
    id: number
    wallet_id: number
    amount: string
    type: string
}

export default function TransactionsPage() {
    const [txs, setTxs] = useState<Transaction[]>([])
    const [loading, setLoading] = useState(true)
    const [isFormOpen, setIsFormOpen] = useState(false)
    const [formData, setFormData] = useState<Transaction>({ id: 0, wallet_id: 0, amount: '0.00', type: 'DEPOSIT' })

    const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const fetchTxs = async () => {
        setLoading(true)
        try {
            const res = await fetch(`${API_URL}/api/transactions`)
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            const list = data.rows ? data.rows.map((row: any[]) => ({
                id: row[0],
                wallet_id: row[1],
                amount: row[2],
                type: row[3]
            })) : []
            setTxs(list)
        } catch (err: any) {
            console.error(err)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchTxs()
    }, [])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        try {
            const res = await fetch(`${API_URL}/api/transactions`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData),
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            setIsFormOpen(false)
            setFormData({ id: 0, wallet_id: 0, amount: '0.00', type: 'DEPOSIT' })
            fetchTxs()
        } catch (err: any) {
            alert(err.message)
        }
    }

    const openCreate = () => {
        setFormData({ id: Math.floor(Math.random() * 100000), wallet_id: 0, amount: '0.00', type: 'DEPOSIT' })
        setIsFormOpen(true)
    }

    return (
        <div className="space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">Transactions</h1>
                    <p className="text-slate-500 mt-1">Transaction history</p>
                </div>
                <Button onClick={openCreate} className="bg-blue-600 hover:bg-blue-700">
                    <Plus size={16} className="mr-2" />
                    New Transaction
                </Button>
            </header>

            {isFormOpen && (
                <Card className="mb-6 border-blue-200 bg-blue-50/50">
                    <CardHeader>
                        <CardTitle>Record Transaction</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="flex gap-4 items-end flex-wrap">
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">ID</label>
                                <input 
                                    type="number" 
                                    className="p-2 border rounded w-24"
                                    value={formData.id}
                                    onChange={e => setFormData({...formData, id: parseInt(e.target.value)})}
                                    required
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Wallet ID</label>
                                <input 
                                    type="number" 
                                    className="p-2 border rounded w-24"
                                    value={formData.wallet_id}
                                    onChange={e => setFormData({...formData, wallet_id: parseInt(e.target.value)})}
                                    required
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Amount</label>
                                <input 
                                    type="text" 
                                    className="p-2 border rounded w-32"
                                    value={formData.amount}
                                    onChange={e => setFormData({...formData, amount: e.target.value})}
                                    required
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">Type</label>
                                <select 
                                    className="p-2 border rounded w-32 bg-white"
                                    value={formData.type}
                                    onChange={e => setFormData({...formData, type: e.target.value})}
                                >
                                    <option value="DEPOSIT">DEPOSIT</option>
                                    <option value="WITHDRAWAL">WITHDRAWAL</option>
                                    <option value="TRANSFER">TRANSFER</option>
                                </select>
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit">Create</Button>
                                <Button type="button" variant="outline" onClick={() => setIsFormOpen(false)}>Cancel</Button>
                            </div>
                        </form>
                    </CardContent>
                </Card>
            )}

            <Card>
                <CardContent className="p-0">
                    <table className="w-full text-sm text-left">
                        <thead className="bg-slate-50 text-slate-500 font-medium">
                            <tr>
                                <th className="px-4 py-3 border-b">ID</th>
                                <th className="px-4 py-3 border-b">Wallet ID</th>
                                <th className="px-4 py-3 border-b">Type</th>
                                <th className="px-4 py-3 border-b">Amount</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr><td colSpan={4} className="p-4 text-center">Loading...</td></tr>
                            ) : txs.length === 0 ? (
                                <tr><td colSpan={4} className="p-4 text-center text-slate-400">No transactions found</td></tr>
                            ) : (
                                txs.map(t => (
                                    <tr key={t.id} className="border-b hover:bg-slate-50">
                                        <td className="px-4 py-3">{t.id}</td>
                                        <td className="px-4 py-3">{t.wallet_id}</td>
                                        <td className="px-4 py-3 text-xs font-semibold px-2 py-1 rounded bg-slate-100 w-min">{t.type}</td>
                                        <td className="px-4 py-3 font-mono">{t.amount}</td>
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
