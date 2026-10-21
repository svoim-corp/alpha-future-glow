import { baseApi } from '@shared/api';

export interface CompanyLookup {
	company: {
		inn: string;
		type: 'LLC' | 'IE';
		name: string;
		ogrn: string;
		regDate: string;
		state: string;
		head: string;
	};
	applicantIsHead: boolean;
	/** NONE — заявитель руководитель, MCHD — нужна доверенность */
	authorityRequired: 'NONE' | 'MCHD';
}

export interface AuthorityCheck {
	verified: boolean;
	principal?: string;
	representative?: string;
	validUntil?: string;
	powers?: string[];
	reason?: string;
}

export const verifyCompanyApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		lookupCompany: build.mutation<CompanyLookup, { applicationId: string; inn: string }>({
			query: ({ applicationId, inn }) => ({
				url: `/applications/${applicationId}/company`,
				method: 'POST',
				body: { inn },
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [{ type: 'Application', id: applicationId }],
		}),
		checkAuthority: build.mutation<AuthorityCheck, { applicationId: string; mchdNumber: string }>({
			query: ({ applicationId, mchdNumber }) => ({
				url: `/applications/${applicationId}/authority`,
				method: 'POST',
				body: { mchdNumber },
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [{ type: 'Application', id: applicationId }],
		}),
	}),
});

export const { useLookupCompanyMutation, useCheckAuthorityMutation } = verifyCompanyApi;
