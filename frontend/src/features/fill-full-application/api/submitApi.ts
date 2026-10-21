import { baseApi } from '@shared/api';

export interface MissingField {
	field: string;
	message: string;
}

export interface ValidationResult {
	valid: boolean;
	missing: MissingField[];
}

export interface SubmitResult {
	id: string;
	number: string;
	status: string;
	redirect: string;
}

export interface SubmitArgs {
	applicationId: string;
	version: number;
	idempotencyKey: string;
}

export const submitApi = baseApi.injectEndpoints({
	endpoints: (build) => ({
		consentGuarantorBki: build.mutation<{ given: boolean; at: string }, string>({
			query: (applicationId) => ({
				url: `/applications/${applicationId}/guarantor/consent-bki`,
				method: 'POST',
				body: { given: true },
			}),
			invalidatesTags: (_result, _error, id) => [{ type: 'Application', id }],
		}),
		validateApplication: build.mutation<ValidationResult, string>({
			query: (applicationId) => ({ url: `/applications/${applicationId}/validate`, method: 'POST' }),
		}),
		submitApplication: build.mutation<SubmitResult, SubmitArgs>({
			query: ({ applicationId, version, idempotencyKey }) => ({
				url: `/applications/${applicationId}/submit`,
				method: 'POST',
				body: { truthConfirmed: true },
				headers: { 'If-Match': String(version), 'Idempotency-Key': idempotencyKey },
			}),
			invalidatesTags: (_result, _error, { applicationId }) => [
				{ type: 'Application', id: applicationId },
				{ type: 'Application', id: 'LIST' },
			],
		}),
	}),
});

export const {
	useConsentGuarantorBkiMutation,
	useValidateApplicationMutation,
	useSubmitApplicationMutation,
} = submitApi;
