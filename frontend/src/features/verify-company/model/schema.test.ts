import { innSchema, mchdSchema } from './schema';

describe('innSchema', () => {
	it.each([['7712345678'], ['771234567890'], [' 7712345678 ']])('принимает %s', (inn) => {
		expect(innSchema.safeParse({ inn }).success).toBe(true);
	});

	it.each([[''], ['123'], ['77123456789'], ['77123456ab'], ['7712345678901']])('отклоняет «%s»', (inn) => {
		expect(innSchema.safeParse({ inn }).success).toBe(false);
	});
});

describe('mchdSchema', () => {
	it('требует непустой номер', () => {
		expect(mchdSchema.safeParse({ mchdNumber: 'a1b2' }).success).toBe(true);
		expect(mchdSchema.safeParse({ mchdNumber: '  ' }).success).toBe(false);
	});
});
