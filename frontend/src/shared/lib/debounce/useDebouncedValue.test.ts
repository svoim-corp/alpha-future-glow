import { act, renderHook } from '@testing-library/react';

import { useDebouncedValue } from './useDebouncedValue';

describe('useDebouncedValue', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('сразу отдаёт начальное значение', () => {
		const { result } = renderHook(() => useDebouncedValue('a', 300));

		expect(result.current).toBe('a');
	});

	it('обновляется только после паузы', () => {
		const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
			initialProps: { value: 'a' },
		});

		rerender({ value: 'b' });
		act(() => {
			vi.advanceTimersByTime(299);
		});
		expect(result.current).toBe('a');

		act(() => {
			vi.advanceTimersByTime(1);
		});
		expect(result.current).toBe('b');
	});

	it('частые изменения сбрасывают таймер', () => {
		const { result, rerender } = renderHook(({ value }) => useDebouncedValue(value, 300), {
			initialProps: { value: 1 },
		});

		rerender({ value: 2 });
		act(() => {
			vi.advanceTimersByTime(200);
		});
		rerender({ value: 3 });
		act(() => {
			vi.advanceTimersByTime(200);
		});
		expect(result.current).toBe(1);

		act(() => {
			vi.advanceTimersByTime(100);
		});
		expect(result.current).toBe(3);
	});
});
