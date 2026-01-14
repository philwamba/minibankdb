'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Pencil, Trash2, Wallet } from 'lucide-react'

type Wallet = {
    id: number
    user_id: number
    balance: string
}

export default function WalletsPage() {
    const [wallets, setWallets] = useState<Wallet[]>([])
    const [loading, setLoading] = useState(true)
    const [editing, setEditing] = useState<Wallet | null>(null)
    const [isFormOpen, setIsFormOpen] = useState(false)
    const [formData, setFormData] = useState<Wallet>({ id: 0, user_id: 0, balance: '0.00' })

    const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const fetchWallets = async () => {
        setLoading(true)
        try {
            const res = await fetch(`${API_URL}/api/wallets`)
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            const list = data.rows ? data.rows.map((row: any[]) => ({
                id: row[0],
                user_id: row[1],
                balance: row[2]
            })) : []
            setWallets(list)
        } catch (err: any) {
            console.error(err)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchWallets()
    }, [])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        try {
            const method = editing ? 'PUT' : 'POST'
            const res = await fetch(`${API_URL}/api/wallets`, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData),
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            setIsFormOpen(false)
            setEditing(null)
            setFormData({ id: 0, user_id: 0, balance: '0.00' })
            fetchWallets()
        } catch (err: any) {
            alert(err.message)
        }
    }

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure?')) return
        try {
            const res = await fetch(`${API_URL}/api/wallets?id=${id}`, {
                method: 'DELETE',
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            fetchWallets()
        } catch (err: any) {
            alert(err.message)
        }
    }

    const openCreate = () => {
        setEditing(null)
        setFormData({ id: Math.floor(Math.random() * 10000), user_id: 0, balance: '0.00' })
        setIsFormOpen(true)
    }

    return (
        <div className="space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">Wallets</h1>
                    <p className="text-slate-500 mt-1">Manage user wallets</p>
                </div>
                <Button onClick={openCreate} className="bg-blue-600 hover:bg-blue-700">
                    <Plus size={16} className="mr-2" />
                    New Wallet
                </Button>
            </header>

            {isFormOpen && (
                <Card className="mb-6 border-blue-200 bg-blue-50/50">
                    <CardHeader>
                        <CardTitle>{editing ? 'Edit Wallet' : 'Create Wallet'}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="flex gap-4 items-end">
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">ID</label>
                                <input 
                                    type="number" 
                                    className="p-2 border rounded w-24"
                                    value={formData.id}
                                    onChange={e => setFormData({...formData, id: parseInt(e.target.value)})}
                                    disabled={!!editing}
                                    required
                                />
                            </div>
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">User ID</label>
                                <input 
                                    type="number" 
                                    className="p-2 border rounded w-24"
                                    value={formData.user_id}
                                    onChange={e => setFormData({...formData, user_id: parseInt(e.target.value)})}
                                    required
                                />
                            </div>
                            <div className="grid gap-2 flex-1">
                                <label className="text-sm font-medium">Balance</label>
                                <input 
                                    type="text" 
                                    className="p-2 border rounded"
                                    value={formData.balance}
                                    onChange={e => setFormData({...formData, balance: e.target.value})}
                                    required
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit">{editing ? 'Update' : 'Create'}</Button>
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
                                <th className="px-4 py-3 border-b">User ID</th>
                                <th className="px-4 py-3 border-b">Balance</th>
                                <th className="px-4 py-3 border-b text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr><td colSpan={4} className="p-4 text-center">Loading...</td></tr>
                            ) : wallets.length === 0 ? (
                                <tr><td colSpan={4} className="p-4 text-center text-slate-400">No wallets found</td></tr>
                            ) : (
                                wallets.map(w => (
                                    <tr key={w.id} className="border-b hover:bg-slate-50">
                                        <td className="px-4 py-3">{w.id}</td>
                                        <td className="px-4 py-3">{w.user_id}</td>
                                        <td className="px-4 py-3 font-mono custom-number">{w.balance}</td>
                                        <td className="px-4 py-3 text-right">
                                            <Button variant="ghost" size="sm" onClick={() => { setEditing(w); setFormData(w); setIsFormOpen(true); }}>
                                                <Pencil size={14} className="text-slate-500" />
                                            </Button>
                                            <Button variant="ghost" size="sm" onClick={() => handleDelete(w.id)}>
                                                <Trash2 size={14} className="text-red-500" />
                                            </Button>
                                        </td>
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
