/**
 * Helpers puros de la landing.
 *
 * La cuenta atrás del lanzamiento vivía dentro del `onMount` de su componente,
 * donde no había forma de probarla: aritmética de fechas que se rompe con un
 * milisegundo mal puesto y que nadie nota hasta que el contador va torcido.
 */

/** Fecha de lanzamiento que anuncia la landing (hora local). */
export const LAUNCH_DATE = '2026-10-01T09:00:00';

/** Cuenta atrás ya formateada a dos dígitos, lista para pintar. */
export interface Countdown {
	days: string;
	hours: string;
	mins: string;
	secs: string;
}

function pad(n: number): string {
	return String(n).padStart(2, '0');
}

/**
 * Tiempo que falta entre `now` y `target`. Nunca cuenta hacia atrás: pasada la
 * fecha se queda en ceros en vez de mostrar un negativo.
 */
export function countdownBetween(target: number, now: number): Countdown {
	const diff = Math.max(0, target - now);
	return {
		days: pad(Math.floor(diff / 86400000)),
		hours: pad(Math.floor((diff % 86400000) / 3600000)),
		mins: pad(Math.floor((diff % 3600000) / 60000)),
		secs: pad(Math.floor((diff % 60000) / 1000))
	};
}

/**
 * Serie del gráfico de la sección "Métricas", en miles de dólares.
 *
 * El gráfico anterior arrancaba en la base del lienzo y subía hasta arriba,
 * así que dibujaba un patrimonio multiplicado por diez mientras el titular
 * decía +12,4%. Estas dos series cuadran entre sí y con las cifras que la
 * landing repite en el hero y en el recorrido del producto: el valor de
 * mercado cierra en $248,500 sobre $221,100 de capital invertido, y esos
 * $27,400 de diferencia son el 12,40% que anuncia la tarjeta.
 */
export const METRICS_MONTHS = [
	'ENE',
	'FEB',
	'MAR',
	'ABR',
	'MAY',
	'JUN',
	'JUL',
	'AGO',
	'SEP',
	'OCT',
	'NOV',
	'DIC'
] as const;

export const METRICS_MARKET_VALUE = [
	182.4, 193, 191, 201.5, 204, 213.2, 209.8, 223.4, 226, 235.1, 241.6, 248.5
];

export const METRICS_INVESTED = [180, 184, 188, 194, 196, 202, 206, 210, 212, 216, 218, 221.1];

/**
 * El eje no arranca en cero: con un recorrido del 12% sobre cifras de seis
 * dígitos, un eje completo aplanaría la curva hasta volverla ilegible. El
 * recorte se dice en el pie del gráfico.
 */
export const METRICS_SCALE = { min: 170, max: 255 };

/** Marcas del eje Y, de arriba abajo. */
export const METRICS_TICKS = [240, 220, 200, 180];

export interface ChartBox {
	left: number;
	right: number;
	top: number;
	bottom: number;
}

/** Coordenada vertical de un valor dentro del área de trazado. */
export function chartY(value: number, scale: { min: number; max: number }, box: ChartBox): number {
	const span = scale.max - scale.min;
	const ratio = span === 0 ? 0 : (value - scale.min) / span;
	return box.bottom - ratio * (box.bottom - box.top);
}

/**
 * Serie ya convertida al atributo `points` de un `<polyline>`, repartiendo los
 * valores a lo ancho del área de trazado.
 */
export function chartPoints(
	values: number[],
	scale: { min: number; max: number },
	box: ChartBox
): string {
	const steps = values.length - 1;
	return values
		.map((value, i) => {
			const x = steps <= 0 ? box.left : box.left + ((box.right - box.left) * i) / steps;
			return `${round(x)},${round(chartY(value, scale, box))}`;
		})
		.join(' ');
}

function round(n: number): number {
	return Math.round(n * 100) / 100;
}
