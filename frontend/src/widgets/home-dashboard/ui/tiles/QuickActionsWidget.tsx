import { type QuickAction } from '@entities/dashboard';
import { ROUTES } from '@shared/config';
import { Icon, type IconName } from '@shared/ui';
import { Link } from 'react-router';

import styles from '../Dashboard.module.scss';

const actions: Record<QuickAction, { title: string; to: string; icon: IconName }> = {
	apply: { title: 'Подать заявку', to: ROUTES.products, icon: 'catalog' },
	plan: { title: 'Планер', to: ROUTES.planner, icon: 'calendar' },
	credits: { title: 'Мои кредиты', to: ROUTES.credits, icon: 'loans' },
	calendar: { title: 'Платежи', to: ROUTES.creditsCalendar, icon: 'calendar' },
	support: { title: 'Написать банку', to: ROUTES.support, icon: 'support' },
};
export const QuickActionsWidget = ({ items }: { items: QuickAction[] }) => <div className={styles.quickActions}>
	{items.map((id) => <Link to={actions[id].to} key={id}><span><Icon name={actions[id].icon} size={20} /></span>{actions[id].title}</Link>)}
</div>;
