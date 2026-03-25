import { useEffect, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../api/client';
import type { Product } from '../types';
import { useCart } from '../context/CartContext';

export function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { addItem } = useCart();
  const [product, setProduct] = useState<Product | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [added, setAdded] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!id) return;
    api.get<Product>(`/products/${id}`)
      .then(setProduct)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const handleAddToCart = () => {
    if (product) {
      addItem(product, 1);
      setAdded(true);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setAdded(false), 2000);
    }
  };

  if (loading) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-gray-500">Loading...</div>
    </div>
  );

  if (error || !product) return (
    <div className="p-8 text-center">
      <p className="text-red-600 mb-4">{error || 'Product not found'}</p>
      <button onClick={() => navigate('/products')} className="text-indigo-600 hover:underline">
        Back to products
      </button>
    </div>
  );

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <button
        onClick={() => navigate('/products')}
        className="text-indigo-600 hover:underline text-sm mb-6 flex items-center gap-1"
      >
        ← Back to products
      </button>

      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="md:flex">
          <div className="md:w-1/2">
            {product.imageUrl ? (
              <img src={product.imageUrl} alt={product.name} className="w-full h-72 md:h-full object-cover" />
            ) : (
              <div className="w-full h-72 md:h-full bg-gradient-to-br from-indigo-100 to-purple-100 flex items-center justify-center">
                <span className="text-8xl">📦</span>
              </div>
            )}
          </div>

          <div className="md:w-1/2 p-8 flex flex-col">
            <div className="flex items-start justify-between mb-2">
              <h1 className="text-2xl font-bold text-gray-900">{product.name}</h1>
              <span className={`text-xs px-2 py-1 rounded-full font-medium ml-2 ${
                product.status === 'published' ? 'bg-green-100 text-green-700' :
                product.status === 'draft' ? 'bg-yellow-100 text-yellow-700' :
                'bg-gray-100 text-gray-600'
              }`}>
                {product.status}
              </span>
            </div>

            <p className="text-sm text-gray-500 mb-4">SKU: {product.sku}</p>

            {product.description && (
              <p className="text-gray-600 mb-6">{product.description}</p>
            )}

            <div className="mt-auto">
              <div className="flex items-center justify-between mb-4">
                <span className="text-3xl font-bold text-indigo-600">${product.price.toFixed(2)}</span>
                <span className={`text-sm px-3 py-1 rounded-full font-medium ${
                  product.stock > 0 ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
                }`}>
                  {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
                </span>
              </div>

              <button
                onClick={handleAddToCart}
                disabled={product.stock === 0}
                className="w-full bg-indigo-600 text-white py-3 rounded-lg font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {added ? 'Added!' : 'Add to Cart'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
