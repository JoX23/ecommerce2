import { Link } from 'react-router-dom';
import type { Product } from '../types';
import { useCart } from '../context/CartContext';

interface Props {
  product: Product;
}

export function ProductCard({ product }: Props) {
  const { addItem } = useCart();

  return (
    <div className="bg-white rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow flex flex-col">
      <Link to={`/products/${product.id}`} className="block">
        {product.imageUrl ? (
          <img
            src={product.imageUrl}
            alt={product.name}
            className="w-full h-48 object-cover rounded-t-lg"
          />
        ) : (
          <div className="w-full h-48 bg-gradient-to-br from-indigo-100 to-purple-100 rounded-t-lg flex items-center justify-center">
            <span className="text-5xl">📦</span>
          </div>
        )}
      </Link>

      <div className="p-4 flex flex-col flex-1">
        <Link to={`/products/${product.id}`}>
          <h3 className="font-semibold text-gray-900 hover:text-indigo-600 truncate">{product.name}</h3>
        </Link>

        {product.description && (
          <p className="text-gray-500 text-sm mt-1 line-clamp-2">{product.description}</p>
        )}

        <div className="mt-auto pt-3 flex items-center justify-between">
          <span className="text-lg font-bold text-indigo-600">${product.price.toFixed(2)}</span>
          <span className={`text-xs px-2 py-1 rounded-full font-medium ${
            product.stock > 0
              ? 'bg-green-100 text-green-700'
              : 'bg-red-100 text-red-700'
          }`}>
            {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
          </span>
        </div>

        <button
          onClick={() => addItem(product)}
          disabled={product.stock === 0}
          className="mt-3 w-full bg-indigo-600 text-white py-2 px-4 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          Add to Cart
        </button>
      </div>
    </div>
  );
}
