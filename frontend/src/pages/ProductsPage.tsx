import { useEffect, useState } from 'react';
import { api } from '../api/client';
import type { Product } from '../types';
import { ProductCard } from '../components/ProductCard';

export function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.get<Product[]>('/products')
      .then(data => setProducts(data ?? []))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-gray-500 text-lg">Loading products...</div>
    </div>
  );

  if (error) return (
    <div className="p-8 text-center text-red-600">{error}</div>
  );

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Products</h1>
      {products.length === 0 ? (
        <div className="text-center text-gray-500 py-16">
          <div className="text-5xl mb-4">📦</div>
          <p className="text-lg">No published products yet.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
          {products.map(p => <ProductCard key={p.id} product={p} />)}
        </div>
      )}
    </div>
  );
}
