/*
 * Los datos de la maqueta de `/dashboard/reports`.
 *
 * Viven aparte de `product-tour.ts` porque son la mitad de sus cifras —la
 * matriz de dos años, la cabecera y las medidas de riesgo— y entre ellas se
 * perdían el menú y las cuatro vistas. Lo que cuentan, en cambio, es lo mismo:
 * el +9,8% de la cabecera sale de encadenar los dos años de la matriz, y es la
 * «rentabilidad real» que enseña el resumen.
 */

/** Abreviaturas de los meses, como en la matriz del panel. */
export const TOUR_MONTHS = [
	'Ene',
	'Feb',
	'Mar',
	'Abr',
	'May',
	'Jun',
	'Jul',
	'Ago',
	'Sep',
	'Oct',
	'Nov',
	'Dic'
];

/**
 * La cifra de cabecera de la ficha de reportes: lo que rindió el dinero en todo
 * el historial.
 *
 * Sale de encadenar los dos años de la matriz de abajo, y es exactamente la
 * «rentabilidad real» que enseña el resumen: las dos pantallas miden lo mismo.
 *
 * No es el +12,40% del titular del patrimonio, y ahí está la gracia: aquel
 * divide la ganancia de hoy entre lo invertido hoy, así que un aporte grande
 * hecho después de una subida lo mueve sin que la cartera haya hecho nada. Las
 * dos cifras coincidían por descuido, que es justo lo que el panel se esfuerza
 * en distinguir.
 */
export const TOUR_RECORD = {
	label: 'Lo que rindió tu dinero',
	value: '+9,8%',
	span: 'Del 1 de enero de 2025 al 30 de noviembre de 2026, 23 meses de historial. Anualizado, eso es un +5,0%.',
	/* El patrimonio y la ganancia son los del resumen: es la misma cuenta. El
	   porcentaje va con un decimal, como lo escribe la cabecera del panel. */
	money:
		'Hoy la cuenta vale $248,500.00 sobre los $221,100.00 que has puesto: +$27,400.00 (+12,4%).'
};

/**
 * Rentabilidad mes a mes, un año por fila y el total cerrando cada uno.
 * Diciembre de 2026 va sin dato porque el año sigue en curso.
 */
export const TOUR_RETURNS = [
	{
		year: '2026',
		values: [2.4, 1.1, -0.8, 1.6, 0.6, 1.2, -1.4, 2.6, -2.0, -1.1, -1.0, null],
		total: 3.1
	},
	{
		year: '2025',
		values: [0.8, 1.2, -1.1, 1.9, 0.5, -0.7, 1.4, 0.6, -1.8, 1.5, 0.9, 1.2],
		total: 6.5
	}
];

/**
 * Las medidas de «Cómo se movió», sacadas de la matriz de arriba: el mejor y el
 * peor mes son suyos, y la máxima caída cubre el tramo de septiembre a noviembre
 * de 2026, que encadena un -4,05%.
 *
 * `hint` no es una ayuda opcional: en el panel es una columna de la tabla,
 * porque «Máxima caída -4,6%» sin ella es un número sin idioma.
 */
export interface TourKeyStat {
	label: string;
	/** Segunda línea bajo la cifra: a qué mes se refiere. */
	detail: string;
	value: string;
	/** Signo con el que teñir la cifra; vacío, no se tiñe. */
	tone: string;
	hint: string;
	/** Reparo que la cifra arrastra y que hay que leer con ella. */
	note?: string;
}

export const TOUR_KEY_STATS: TourKeyStat[] = [
	{
		label: 'Mejor mes',
		detail: 'agosto de 2026',
		value: '+2,6%',
		tone: 'up',
		hint: 'El mes que más rindió; los incompletos no compiten.'
	},
	{
		label: 'Peor mes',
		detail: 'septiembre de 2026',
		value: '-2,0%',
		tone: 'down',
		hint: 'El mes que menos rindió, con la misma regla.'
	},
	{
		label: 'Volatilidad anualizada',
		detail: '',
		value: '4,3%',
		tone: '',
		hint: 'Cuánto oscila tu rentabilidad de un día al siguiente, llevada a un año.'
	},
	{
		label: 'Máxima caída',
		detail: '',
		value: '-4,6%',
		tone: 'down',
		hint: 'La peor bajada desde un máximo hasta el siguiente suelo.'
	},
	{
		label: 'Ratio de Sharpe',
		detail: '',
		/* Un cociente, no un porcentaje, y con dos decimales: es como lo escribe
		   la ficha. Sale del +5,0% anualizado sobre el 4,3% de volatilidad. */
		value: '1,16',
		/* En gris siempre: es una estimación, no una ganancia, y pintarla en verde
		   la vendería como un sello de calidad. */
		tone: '',
		hint: 'Cuánta rentabilidad te dio cada unidad de riesgo, con tasa libre de riesgo 0.',
		note: 'Estimación: con pocos meses de historial su margen de error es amplio.'
	}
];

/**
 * Fondo de una celda de la matriz: verde lo que subió, rojo lo que bajó, y más
 * intenso cuanto mayor fue el movimiento.
 *
 * Espeja `returnBackground` de `features/reports`, con la misma escala y el
 * mismo punto de saturación. Está copiado y no importado porque una feature no
 * importa de otra (docs/FRONTEND_ARCHITECTURE.md); si allí cambia la escala,
 * aquí también.
 */
export function tourReturnBackground(value: number | null): string {
	if (value === null) return '';

	const intensity = Math.min(Math.abs(value) / 2.5, 1);
	const alpha = (0.05 + intensity * 0.2).toFixed(3);

	return value < 0 ? `rgba(224, 90, 90, ${alpha})` : `rgba(34, 201, 126, ${alpha})`;
}
