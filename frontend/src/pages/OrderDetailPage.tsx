import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { Order } from '../types';

const STATUS_COLORS: Record<string, string> = {
  pending:   'bg-yellow-100 text-yellow-700',
  confirmed: 'bg-blue-100 text-blue-700',
  shipped:   'bg-indigo-100 text-indigo-700',
  delivered: 'bg-green-100 text-green-700',
  cancelled: 'bg-red-100 text-red-700',
};

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!id) return;
    api.get<Order>(`/orders/${id}`)
      .then(setOrder)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-gray-500">Loading order...</div>
    </div>
  );

  if (error || !order) return (
    <div className="p-8 text-center">
      <p className="text-red-600 mb-4">{error || 'Order not found'}</p>
      <button onClick={() => navigate('/orders')} className="text-indigo-600 hover:underline">
        Back to orders
      </button>
    </div>
  );

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <button
        onClick={() => navigate('/orders')}
        className="text-indigo-600 hover:underline text-sm mb-6 flex items-center gap-1"
      >
        ← Back to orders
      </button>

      <div className="bg-white rounded-xl border border-gray-200 shadow-sm p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-bold text-gray-900">Order Details</h1>
            <p className="text-sm text-gray-500 font-mono mt-1">#{order.id}</p>
          </div>
          <span className={`px-3 py-1 rounded-full text-sm font-medium ${STATUS_COLORS[order.status] ?? 'bg-gray-100 text-gray-600'}`}>
            {order.status}
          </span>
        </div>

        <div className="text-sm text-gray-500 mb-6">
          Placed on {new Date(order.createdAt).toLocaleString()}
        </div>

        <h2 className="font-semibold text-gray-900 mb-3">Items</h2>
        <div className="space-y-3 mb-6">
          {order.items?.map((item, i) => (
            <div key={i} className="flex items-center justify-between py-3 border-b border-gray-100 last:border-0">
              <div>
                <p className="text-sm font-medium text-gray-900 font-mono">
                  {item.productId.slice(0, 8)}...
                </p>
                <p className="text-xs text-gray-500">
                  {item.qty} × ${item.unitPrice.toFixed(2)}
                </p>
              </div>
              <span className="font-semibold text-gray-900">${item.subtotal.toFixed(2)}</span>
            </div>
          ))}
        </div>

        <div className="flex items-center justify-between pt-4 border-t border-gray-200">
          <span className="text-lg font-semibold text-gray-900">Total</span>
          <span className="text-2xl font-bold text-indigo-600">${order.total.toFixed(2)}</span>
        </div>
      </div>
    </div>
  );
}
