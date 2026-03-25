import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCart } from '../context/CartContext';
import { api } from '../api/client';
import type { Order } from '../types';

export function CartPage() {
  const { items, removeItem, updateQty, clearCart, totalPrice } = useCart();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleCheckout = async () => {
    if (items.length === 0) return;
    setError('');
    setLoading(true);
    try {
      await api.post<Order>('/orders', {
        items: items.map(i => ({ productId: i.product.id, qty: i.qty })),
      });
      clearCart();
      navigate('/orders');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Checkout failed');
    } finally {
      setLoading(false);
    }
  };

  if (items.length === 0) return (
    <div className="max-w-2xl mx-auto px-4 py-16 text-center">
      <div className="text-6xl mb-4">🛒</div>
      <h1 className="text-2xl font-bold text-gray-900 mb-2">Your cart is empty</h1>
      <p className="text-gray-500 mb-6">Add some products to get started.</p>
      <button
        onClick={() => navigate('/products')}
        className="bg-indigo-600 text-white px-6 py-2 rounded-lg hover:bg-indigo-700"
      >
        Browse Products
      </button>
    </div>
  );

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Your Cart</h1>

      {error && (
        <div className="mb-4 bg-red-50 border border-red-200 text-red-700 rounded-lg px-4 py-3 text-sm">
          {error}
        </div>
      )}

      <div className="space-y-4 mb-8">
        {items.map(item => (
          <div key={item.product.id} className="bg-white rounded-lg border border-gray-200 p-4 flex items-center gap-4">
            <div className="w-16 h-16 bg-indigo-50 rounded-lg flex items-center justify-center flex-shrink-0">
              {item.product.imageUrl ? (
                <img src={item.product.imageUrl} alt={item.product.name} className="w-full h-full object-cover rounded-lg" />
              ) : (
                <span className="text-2xl">📦</span>
              )}
            </div>

            <div className="flex-1 min-w-0">
              <h3 className="font-semibold text-gray-900 truncate">{item.product.name}</h3>
              <p className="text-sm text-gray-500">${item.product.price.toFixed(2)} each</p>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => updateQty(item.product.id, item.qty - 1)}
                className="w-7 h-7 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 text-lg font-medium"
              >
                −
              </button>
              <span className="w-8 text-center font-medium">{item.qty}</span>
              <button
                onClick={() => updateQty(item.product.id, item.qty + 1)}
                className="w-7 h-7 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 text-lg font-medium"
              >
                +
              </button>
            </div>

            <div className="text-right min-w-[80px]">
              <p className="font-semibold text-indigo-600">${(item.product.price * item.qty).toFixed(2)}</p>
            </div>

            <button
              onClick={() => removeItem(item.product.id)}
              className="text-red-400 hover:text-red-600 text-lg ml-2"
            >
              ✕
            </button>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-4">
          <span className="text-lg font-semibold text-gray-900">Total</span>
          <span className="text-2xl font-bold text-indigo-600">${totalPrice.toFixed(2)}</span>
        </div>
        <button
          onClick={handleCheckout}
          disabled={loading}
          className="w-full bg-indigo-600 text-white py-3 rounded-lg font-medium hover:bg-indigo-700 disabled:opacity-50 transition-colors"
        >
          {loading ? 'Placing order...' : 'Place Order'}
        </button>
      </div>
    </div>
  );
}
