/*
 * Los iconos de la maqueta del panel, como datos.
 *
 * Son los trazos que usa el menú lateral de `product-tour-window.svelte`,
 * copiados de `dashboard/icons.ts` porque una feature no importa de otra
 * (docs/FRONTEND_ARCHITECTURE.md). Viven aparte de `product-tour.ts` por la
 * misma razón que allí: quince bloques de trazos escondían los datos del
 * recorrido entre ellos.
 *
 * Todos son de 24×24 sin relleno, para que hereden grosor y color de quien los
 * pinta en vez de traerlos puestos.
 */
export const TOUR_ICONS: Record<string, string[]> = {
	grid: ['M3 3h7v7H3zM14 3h7v7h-7zM14 14h7v7h-7zM3 14h7v7H3z'],
	briefcase: [
		'M4 7h16a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2z',
		'M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2'
	],
	pie: ['M21.21 15.89A10 10 0 1 1 8 2.83', 'M22 12A10 10 0 0 0 12 2v10z'],
	layers: ['M12 2l8 4v12l-8 4-8-4V6l8-4z', 'M12 22V12', 'M4 6l8 6 8-6', 'M4 18l8-6 8 6'],
	exchange: ['M12 5v14', 'M19 12l-7 7-7-7'],
	bars: ['M12 20V10', 'M18 20V4', 'M6 20v-4'],
	bell: ['M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9', 'M13.73 21a2 2 0 0 1-3.46 0'],
	book: [
		'M4 19.5A2.5 2.5 0 0 1 6.5 17H20',
		'M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z'
	],
	gear: [
		'M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z',
		'M12 1v6m0 6v6M4.22 4.22l4.24 4.24m5.08 5.08l4.24 4.24M1 12h6m6 0h6m-1.78 7.78l-4.24-4.24m-5.08-5.08l-4.24-4.24'
	],
	logout: ['M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4', 'M16 17l5-5-5-5', 'M21 12H9'],
	eye: ['M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z', 'M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0z']
};
