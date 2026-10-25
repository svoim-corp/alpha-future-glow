import { render, screen } from '@testing-library/react';

import { ChatBubble } from './ChatBubble';

it('обозначает автора и безопасно выводит текст сообщения', () => {
	render(<ChatBubble speaker="assistant"><p>{'<script>alert(1)</script>'}</p></ChatBubble>);
	expect(screen.getByRole('article', { name: 'Ассистент' })).toBeInTheDocument();
	expect(screen.getByText('<script>alert(1)</script>')).toBeInTheDocument();
});
