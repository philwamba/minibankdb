'use client'

import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus, Pencil, Trash2 } from 'lucide-react'

type User = {
    id: number
    name: string
    email: string
}

export default function UsersPage() {
    const [users, setUsers] = useState<User[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [editingUser, setEditingUser] = useState<User | null>(null)
    const [isFormOpen, setIsFormOpen] = useState(false)

    // Form state
    const [formData, setFormData] = useState<User>({ id: 0, name: '', email: '' })
    
    const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    const fetchUsers = async () => {
        setLoading(true)
        try {
            const res = await fetch(`${API_URL}/api/users`)
            if (!res.ok) throw new Error(`HTTP ${res.status}`)
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            const usersList = data.rows ? data.rows.map((row: any[]) => ({
                id: row[0],
                name: row[1],
                email: row[2]
            })) : []
            setUsers(usersList)
        } catch (err: any) {
            setError(err.message)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        fetchUsers()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        try {
            const method = editingUser ? 'PUT' : 'POST'
            const res = await fetch(`${API_URL}/api/users`, {
                method,
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(formData),
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            
            setIsFormOpen(false)
            setEditingUser(null)
            setFormData({ id: 0, name: '', email: '' })
            fetchUsers()
        } catch (err: any) {
            alert(err.message)
        }
    }

    const handleDelete = async (id: number) => {
        if (!confirm('Are you sure?')) return
        try {
            const res = await fetch(`${API_URL}/api/users?id=${id}`, {
                method: 'DELETE',
            })
            const data = await res.json()
            if (data.error) throw new Error(data.error)
            fetchUsers()
        } catch (err: any) {
            alert(err.message)
        }
    }

    const openCreate = () => {
        setEditingUser(null)
        setFormData({ id: Math.floor(Math.random() * 100000000), name: '', email: '' })
        setIsFormOpen(true)
    }

    const openEdit = (user: User) => {
        setEditingUser(user)
        setFormData(user)
        setIsFormOpen(true)
    }

    return (
        <div className="space-y-6">
            <header className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold text-slate-900">Users</h1>
                    <p className="text-slate-500 mt-1">Manage bank users</p>
                </div>
                <Button onClick={openCreate} className="bg-blue-600 hover:bg-blue-700">
                    <Plus size={16} className="mr-2" />
                    New User
                </Button>
            </header>

            {isFormOpen && (
                <Card className="mb-6 border-blue-200 bg-blue-50/50">
                    <CardHeader>
                        <CardTitle>{editingUser ? 'Edit User' : 'Create User'}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="flex gap-4 items-end">
                            <div className="grid gap-2">
                                <label className="text-sm font-medium">ID</label>
                                <input 
                                    type="number" 
                                    className="p-2 border rounded"
                                    value={formData.id}
                                    onChange={e => setFormData({...formData, id: parseInt(e.target.value)})}
                                    disabled={!!editingUser}
                                    required
                                />
                            </div>
                            <div className="grid gap-2 flex-1">
                                <label className="text-sm font-medium">Name</label>
                                <input 
                                    type="text" 
                                    className="p-2 border rounded"
                                    value={formData.name}
                                    onChange={e => setFormData({...formData, name: e.target.value})}
                                    required
                                />
                            </div>
                            <div className="grid gap-2 flex-1">
                                <label className="text-sm font-medium">Email</label>
                                <input 
                                    type="email" 
                                    className="p-2 border rounded"
                                    value={formData.email}
                                    onChange={e => setFormData({...formData, email: e.target.value})}
                                    required
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button type="submit">{editingUser ? 'Update' : 'Create'}</Button>
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
                                <th className="px-4 py-3 border-b">Name</th>
                                <th className="px-4 py-3 border-b">Email</th>
                                <th className="px-4 py-3 border-b text-right">Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr><td colSpan={4} className="p-4 text-center">Loading...</td></tr>
                            ) : users.length === 0 ? (
                                <tr><td colSpan={4} className="p-4 text-center text-slate-400">No users found</td></tr>
                            ) : (
                                users.map(user => (
                                    <tr key={user.id} className="border-b hover:bg-slate-50">
                                        <td className="px-4 py-3">{user.id}</td>
                                        <td className="px-4 py-3 font-medium">{user.name}</td>
                                        <td className="px-4 py-3 text-slate-500">{user.email}</td>
                                        <td className="px-4 py-3 text-right">
                                            <Button variant="ghost" size="sm" onClick={() => openEdit(user)}>
                                                <Pencil size={14} className="text-slate-500" />
                                            </Button>
                                            <Button variant="ghost" size="sm" onClick={() => handleDelete(user.id)}>
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
