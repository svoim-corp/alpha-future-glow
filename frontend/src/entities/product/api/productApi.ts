import { baseApi } from '@shared/api';

import { type Goal, type MatchRequest, type ProductDetails, type ProductMatch, type ProductSummary } from '../model/types';

export const productApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		getProducts: build.query<ProductSummary[], void>({
			query: () => '/products',
			providesTags: ['Product'],
		}),
		getProduct: build.query<ProductDetails, string>({
			query: (code) => `/products/${code}`,
			providesTags: (_result, _error, code) => [{ type: 'Product', id: code }],
		}),
		getGoals: build.query<Goal[], void>({
			query: () => '/products/goals',
			providesTags: ['Product'],
		}),
		/** Подбор по задаче: какие продукты подходят под цель, сумму и срок */
		matchProducts: build.query<ProductMatch[], MatchRequest>({
			query: (body) => ({ url: '/products/match', method: 'POST', body }),
		}),
	}),
});

export const { useGetProductsQuery, useGetProductQuery, useGetGoalsQuery, useMatchProductsQuery } = productApi;
