import { type DashboardWidget } from '@entities/dashboard';

import { ApplicationsWidget, ApplicationWidget } from './ApplicationWidgets';
import { NotificationsWidget, TicketsWidget } from './CommunicationWidgets';
import { FundingWidget, PlanFundingWidget } from './FundingWidgets';
import { DebtWidget, LoanWidget, PaymentWidget } from './LoanWidgets';
import { DraftsWidget, OffersWidget } from './OfferWidgets';
import { CalendarWidget, PlansWidget } from './PlanWidgets';
import { LoansWidget, PaymentsWidget } from './PortfolioWidgets';
import { QuickActionsWidget } from './QuickActionsWidget';

export const WidgetContent = ({ widget, onConfigure }: { widget: DashboardWidget; onConfigure: () => void }) => {
	switch (widget.type) {
		case 'applications': return <ApplicationsWidget widget={widget} />;
		case 'loan': return <LoanWidget widget={widget} onConfigure={onConfigure} />;
		case 'payment': return <PaymentWidget widget={widget} />;
		case 'debt': return <DebtWidget />;
		case 'quick': return <QuickActionsWidget items={widget.settings.actions} />;
		case 'calendar': return <CalendarWidget widget={widget} />;
		case 'plans': return <PlansWidget widget={widget} />;
		case 'application': return <ApplicationWidget widget={widget} onConfigure={onConfigure} />;
		case 'loans': return <LoansWidget widget={widget} />;
		case 'payments': return <PaymentsWidget widget={widget} onConfigure={onConfigure} />;
		case 'offers': return <OffersWidget widget={widget} />;
		case 'drafts': return <DraftsWidget widget={widget} />;
		case 'planFunding': return <PlanFundingWidget widget={widget} onConfigure={onConfigure} />;
		case 'notifications': return <NotificationsWidget widget={widget} />;
		case 'tickets': return <TicketsWidget widget={widget} />;
		case 'funding': return <FundingWidget widget={widget} />;
	}
};
