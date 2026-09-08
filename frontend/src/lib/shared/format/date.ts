/**
 * Formats a calendar-only date value ("2026-07-07" or a UTC-midnight ISO
 * timestamp like "2026-07-07T00:00:00Z") without letting the browser's
 * local timezone shift it to the previous or next day.
 */
export function formatCalendarDate(
	value: string,
	options: Intl.DateTimeFormatOptions,
	locale = 'es-CO'
): string {
	const [year, month, day] = value.split('T')[0].split('-').map(Number);
	return new Date(Date.UTC(year, month - 1, day)).toLocaleDateString(locale, {
		...options,
		timeZone: 'UTC'
	});
}

/**
 * Returns "today" as "YYYY-MM-DD" in the browser's local timezone, for use
 * as a default value in date-only form fields. Avoids `toISOString()`,
 * which reflects the UTC date and can be off by a day for the user's
 * local calendar date.
 */
export function todayLocalDateString(): string {
	const now = new Date();
	const month = String(now.getMonth() + 1).padStart(2, '0');
	const day = String(now.getDate()).padStart(2, '0');
	return `${now.getFullYear()}-${month}-${day}`;
}

/**
 * Cuánto hace de un instante: «hace un momento», «hace 3 horas», «hace 2 días».
 *
 * No es `formatCalendarDate`: aquello imprime un día del calendario y esto mide
 * una distancia. La diferencia importa cuando lo que se está juzgando es si un
 * dato sigue valiendo — un precio traído hace diez minutos y otro traído esta
 * mañana caen los dos en «hoy», y solo uno de los dos hay que volver a pedir.
 *
 * Devuelve cadena vacía si no hay fecha, para que quien lo llama decida qué
 * decir en su lugar: «nunca» y «no aplica» no son lo mismo en todas las
 * pantallas.
 */
export function formatTimeAgo(
	iso: string | null | undefined,
	now: Date = new Date(),
	locale = 'es'
): string {
	if (!iso) return '';

	const then = new Date(iso);
	if (Number.isNaN(then.getTime())) return '';

	const seconds = Math.round((now.getTime() - then.getTime()) / 1000);

	// Por debajo del minuto no hay nada que contar, y «hace 40 segundos» se lee
	// como una precisión que el dato no tiene.
	if (seconds < 60) return 'hace un momento';

	const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });

	const units: [Intl.RelativeTimeFormatUnit, number][] = [
		['minute', 60],
		['hour', 3600],
		['day', 86400],
		['month', 2592000],
		['year', 31536000]
	];

	// El último tramo que aún cabe: se recorre de mayor a menor y se para en el
	// primero cuyo valor sea al menos 1.
	for (let i = units.length - 1; i >= 0; i--) {
		const [unit, size] = units[i];
		const value = Math.floor(seconds / size);
		if (value >= 1) return rtf.format(-value, unit);
	}

	return 'hace un momento';
}
