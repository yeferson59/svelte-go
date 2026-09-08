import { describe, it, expect } from 'vitest';
import { formatCalendarDate, formatTimeAgo, todayLocalDateString } from './date';

describe('formatCalendarDate', () => {
	it('keeps the calendar day for a UTC-midnight ISO timestamp regardless of local timezone', () => {
		expect(
			formatCalendarDate('2026-07-07T00:00:00Z', {
				year: 'numeric',
				month: '2-digit',
				day: '2-digit'
			})
		).toBe('07/07/2026');
	});

	it('keeps the calendar day for a plain date-only string', () => {
		expect(
			formatCalendarDate('2026-01-31', { year: 'numeric', month: '2-digit', day: '2-digit' })
		).toBe('31/01/2026');
	});
});

describe('todayLocalDateString', () => {
	it('matches the local Y-M-D components of the current time', () => {
		const now = new Date();
		const expected = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
		expect(todayLocalDateString()).toBe(expected);
	});
});

describe('formatTimeAgo', () => {
	const now = new Date('2026-09-07T12:00:00Z');

	it('does not pretend to a precision the data does not have under a minute', () => {
		expect(formatTimeAgo('2026-09-07T11:59:22Z', now)).toBe('hace un momento');
	});

	// El tramo que importa: un precio traído hace diez minutos y otro traído
	// esta mañana caen los dos en «hoy», y solo uno de los dos hay que volver a
	// pedir.
	it('separates what a calendar day would collapse', () => {
		expect(formatTimeAgo('2026-09-07T11:50:00Z', now)).toBe('hace 10 minutos');
		expect(formatTimeAgo('2026-09-07T05:00:00Z', now)).toBe('hace 7 horas');
	});

	it('climbs to the largest unit that still fits', () => {
		expect(formatTimeAgo('2026-09-04T12:00:00Z', now)).toBe('hace 3 días');
		expect(formatTimeAgo('2026-06-07T12:00:00Z', now)).toBe('hace 3 meses');
		expect(formatTimeAgo('2024-09-07T12:00:00Z', now)).toBe('hace 2 años');
	});

	// Vacío y no «nunca»: quien llama decide qué decir, porque «nunca» y «no
	// aplica» no son lo mismo en todas las pantallas.
	it('returns nothing for a missing or unparseable date', () => {
		expect(formatTimeAgo(null, now)).toBe('');
		expect(formatTimeAgo(undefined, now)).toBe('');
		expect(formatTimeAgo('no es una fecha', now)).toBe('');
	});
});
