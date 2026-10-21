import { useDispatch, useSelector } from 'react-redux';

/**
 * Типы RootState / AppDispatch объявлены глобально в app/model/types.ts,
 * чтобы shared не импортировал app (FSD).
 */
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();
