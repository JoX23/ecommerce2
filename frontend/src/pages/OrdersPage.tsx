import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import type { Order } from '../types';
import { STATUS_COLORS } from '../utils/orderStatus';

export function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.get<{ data: Order[] }>('/orders')
      .then(res => setOrders(res?.data ?? []))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-gray-500">Loading orders...</div>
    </div>
  );

  if (error) return <div className="p-8 text-center text-red-600">{error}</div>;

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Your Orders</h1>

      {orders.length === 0 ? (
        <div className="text-center py-16">
          <div className="text-5xl mb-4">📋</div>
          <p className="text-gray-500 text-lg">No orders yet.</p>
          <Link to="/products" className="mt-4 inline-block text-indigo-600 hover:underline">
            Start shopping
          </Link>
        </div>
      ) : (
        <div className="space-y-4">
          {orders.map(order => (
            <Link
              key={order.id}
              to={`/orders/${order.id}`}
              className="block bg-white rounded-lg border border-gray-200 p-5 hover:shadow-md transition-shadow"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-mono text-sm text-gray-500">#{order.id.slice(0, 8)}</span>
                <span className={`text-xs px-2 py-1 rounded-full font-medium ${STATUS_COLORS[order.status] ?? 'bg-gray-100 text-gray-600'}`}>
                  {order.status}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600">
                  {order.items?.length ?? 0} item{(order.items?.length ?? 0) !== 1 ? 's' : ''}
                  {' · '}
                  {new Date(order.createdAt).toLocaleDateString()}
                </span>
                <span className="font-bold text-indigo-600">${order.total.toFixed(2)}</span>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
