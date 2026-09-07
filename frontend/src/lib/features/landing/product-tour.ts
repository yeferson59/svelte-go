/**
 * Datos del recorrido "Dentro de Finexia" de la landing.
 *
 * Son cifras de ejemplo, no datos reales: viven aquí en vez de dentro de los
 * componentes para que las cuatro maquetas cuadren entre sí (el patrimonio del
 * resumen es la suma de los portafolios y el de las plataformas, y el
 * rendimiento es el mismo en todas las vistas). Si cambia una cifra, cambia en
 * un solo sitio.
 *
 * Los importes van escritos como los escribe la aplicación: `formatCurrency`
 * da a cada moneda el locale en el que se lee, y el dólar se lee en en-US
 * («$248,500.00»). Los porcentajes van en es-CO, con coma. La maqueta decía los
 * importes al revés que el panel que dice retratar.
 */

/**
 * El menú lateral del panel, en el mismo orden y con los mismos iconos.
 *
 * Espeja `MAIN_NAV` de `features/dashboard`. Está copiado y no importado porque
 * una feature no importa de otra (docs/FRONTEND_ARCHITECTURE.md); si allí se
 * añade o se renombra una sección, aquí también. «Inversiones» no sale porque
 * su feature flag viene apagada.
 */
export const TOUR_NAV = [
	{ label: 'Resumen', icon: 'grid' },
	{ label: 'Portafolios', icon: 'briefcase' },
	{ label: 'Mis activos', icon: 'pie' },
	{ label: 'Plataformas', icon: 'layers' },
	{ label: 'Transacciones', icon: 'exchange' },
	{ label: 'Reportes', icon: 'bars' },
	{ label: 'Notificaciones', icon: 'bell' },
	{ label: 'Guía de usuario', icon: 'book' },
	{ label: 'Configuración', icon: 'gear' }
] as const;

export type TourNavLabel = (typeof TOUR_NAV)[number]['label'];

export { TOUR_ICONS } from './tour-icons';

/** Quien aparece en la barra superior de la maqueta. Inventada, y se nota. */
export const TOUR_USER = {
	name: 'Laura Méndez',
	email: 'laura@ejemplo.com',
	initial: 'L'
};

export type TourViewId = 'resumen' | 'portafolios' | 'transacciones' | 'reportes';

export interface TourView {
	id: TourViewId;
	/** Etiqueta de la pestaña. */
	tab: string;
	/**
	 * Sección del menú que queda marcada. Es también lo que dice la barra
	 * superior: en el panel las dos salen de la misma lista, así que no pueden
	 * discrepar.
	 */
	nav: TourNavLabel;
	/** Ruta de la vista, para la barra de direcciones de la maqueta. */
	path: string;
	/** Titular del pie que describe la vista. */
	title: string;
	/** Qué resuelve la vista, en una frase. */
	description: string;
	/** Capacidades concretas, como etiquetas. */
	points: string[];
}

export const TOUR_VIEWS: TourView[] = [
	{
		id: 'resumen',
		tab: 'Resumen',
		nav: 'Resumen',
		path: '/dashboard',
		title: 'Tu patrimonio agregado, en una pantalla',
		description:
			'El total de todos tus portafolios, dónde está repartido y cómo ha evolucionado sobre el capital que pusiste. Una cifra grande, y debajo la misma cuenta desglosada.',
		points: [
			'Reparto por plataforma, portafolio o tipo',
			'Valor de mercado frente a capital invertido',
			'1M · 3M · 6M · 1Y · Todo',
			'Once monedas de visualización',
			'Ocultar los importes'
		]
	},
	{
		id: 'portafolios',
		tab: 'Portafolios',
		nav: 'Portafolios',
		path: '/dashboard/portfolios',
		title: 'Los portafolios que tú defines',
		description:
			'Agrupa activos de plataformas distintas bajo el portafolio que tengas en mente. Cada fila enseña su perfil de riesgo y, en una barra, cuánto pusiste y qué ha hecho el mercado con ello.',
		points: [
			'Perfil de riesgo por portafolio',
			'Capital y ganancia en una barra',
			'Rendimiento por portafolio',
			'Divisa propia por portafolio',
			'El total, al pie de su columna'
		]
	},
	{
		id: 'transacciones',
		tab: 'Transacciones',
		nav: 'Transacciones',
		path: '/dashboard/transactions',
		title: 'Cada movimiento, registrado por ti',
		description:
			'Compras, ventas, dividendos, intereses, traspasos y cargos, del más reciente al más antiguo. Uno a uno desde la posición del activo, o subiendo el extracto de tu bróker de una vez.',
		points: [
			'Ocho tipos de movimiento',
			'Cantidad, precio y total',
			'Tu nota en cada apunte',
			'Importación desde Excel',
			'Conversión con la tasa del día'
		]
	},
	{
		id: 'reportes',
		tab: 'Reportes',
		nav: 'Reportes',
		path: '/dashboard/reports',
		title: 'Reportes que puedes llevarte',
		description:
			'Qué rindió tu dinero mes a mes y con cuánto vaivén, con lo que mide cada cifra escrito al lado. Todo descargable en XLSX, porque los datos son tuyos.',
		points: [
			'Rentabilidad mes a mes',
			'Volatilidad, máxima caída y Sharpe',
			'Qué mide cada medida, en la tabla',
			'Proyección a cinco años',
			'Descarga en XLSX'
		]
	}
];

/* ---------------------------------------------------------------------------
 * Resumen (`/dashboard`)
 * ------------------------------------------------------------------------- */

/** La cifra grande y las dos líneas que la acompañan. */
export const TOUR_NET_WORTH = {
	label: 'Patrimonio total',
	total: '$248,500.00',
	delta: '+$27,400.00 sobre lo invertido (+12,40%)',
	meta: 'Repartido en 3 portafolios y 18 posiciones.',
	currency: 'USD',
	since: 'Desde enero de 2025'
};

/**
 * Las dos series de la gráfica, en miles de dólares.
 *
 * El valor de mercado y el capital invertido, que es el par que dibuja el
 * panel. El capital sube a escalones porque solo cambia cuando entra dinero, y
 * la distancia entre las dos al final son los $27,400 de ganancia del titular.
 *
 * Estaban en porcentaje de altura del lienzo, así que el trazo cerraba donde
 * cayera: el último punto valía unos $242,000 mientras la cifra de arriba decía
 * $248,500. Escritas en dinero, la gráfica no puede contradecir al titular.
 */
export const TOUR_GROWTH_MARKET = [
	178, 185, 182, 193, 199, 196, 207, 214, 211, 225, 234, 242, 248.5
];
export const TOUR_GROWTH_COST = [170, 170, 180, 180, 180, 191, 191, 199, 199, 208, 208, 215, 221.1];

/**
 * El eje vertical, en miles: dominio y paso entre marcas.
 *
 * Los $20,000 de paso son los que sacaría `growthScale` para este recorrido
 * —redondea el paso al 1, 2, 5 o 10 de su magnitud—, y dan las seis marcas que
 * enseña el panel. La maqueta ponía tres, repartidas a ojo y sin coincidir con
 * sus líneas.
 */
export const TOUR_GROWTH_SCALE = { min: 160, max: 260, step: 20 };

/** Marcas del eje horizontal: el primer punto, el del medio y el último. */
export const TOUR_GROWTH_DATES = ['ene de 25', 'dic de 25', 'nov de 26'];

/** Las cuatro medidas que van sobre la gráfica. */
export const TOUR_GROWTH_STATS = [
	{ label: 'Total ganancia', value: '+$27,400.00', tone: 'up' },
	{ label: 'Rendimiento', value: '+12,40%', tone: 'up' },
	/*
	 * No es el rendimiento de al lado: encadena tramos y no se hunde con un
	 * aporte hecho después de una subida. Por eso son distintas —y por eso es
	 * exactamente el «lo que rindió tu dinero» de la ficha de reportes, que mide
	 * lo mismo—. Con un decimal, como la escribe el panel.
	 */
	{ label: 'Rentabilidad real, Todo', value: '+9,8%', tone: 'up' },
	{ label: 'Valor actual en USD', value: '$248,500.00', tone: 'amber' }
];

/** Los tres cortes de «Dónde está»; abre por el primero que tenga filas. */
export const TOUR_CUTS = ['Plataforma', 'Portafolio', 'Tipo de activo'];

/**
 * El reparto por plataforma.
 *
 * Son las cinco que nombra el mapa del hero, y de aquí las lee: eran tres con
 * otros nombres, así que la portada prometía «5 plataformas · 3 portafolios» y
 * dos pantallas más abajo el reparto solo enseñaba tres, con otras marcas.
 *
 * Los valores suman el patrimonio y los costes implícitos suman el capital
 * invertido, así que el rendimiento ponderado de las cinco filas es el +12,40%
 * del titular; las posiciones suman las 18 que anuncia.
 */
export const TOUR_BREAKDOWN = [
	{
		name: 'Degiro',
		detail: '6 posiciones abiertas',
		share: 33.88,
		value: '$84,200.00',
		gain: '+16,78%',
		up: true
	},
	{
		name: 'Binance',
		detail: '4 posiciones abiertas',
		share: 22.49,
		value: '$55,900.00',
		gain: '+24,22%',
		up: true
	},
	{
		name: 'Revolut',
		detail: '3 posiciones abiertas',
		share: 16.62,
		value: '$41,300.00',
		gain: '+1,72%',
		up: true
	},
	{
		name: 'IBKR',
		detail: '3 posiciones abiertas',
		share: 16.06,
		value: '$39,900.00',
		gain: '+13,68%',
		up: true
	},
	{
		name: 'Banco',
		detail: '2 posiciones abiertas',
		share: 10.95,
		value: '$27,200.00',
		gain: '-3,89%',
		up: false
	}
];

/** El extracto de la derecha: los últimos movimientos, con su signo. */
export const TOUR_ACTIVITY = [
	{
		kind: 'Dividendo',
		asset: 'Apple Inc. (AAPL)',
		amount: '+$10.92',
		date: '18 de jun',
		incoming: true
	},
	{
		kind: 'Compra',
		asset: 'Vanguard FTSE All-World UCITS ETF (VWCE)',
		amount: '−$1,896.00',
		date: '11 de jun',
		incoming: false
	},
	{
		kind: 'Interés',
		asset: 'Efectivo en dólares (USD)',
		amount: '+$19.95',
		date: '2 de jun',
		incoming: true
	},
	{
		kind: 'Venta',
		asset: 'Ethereum (ETH)',
		amount: '−$2,046.00',
		date: '27 de may',
		incoming: false
	}
];

/* ---------------------------------------------------------------------------
 * Portafolios (`/dashboard/portfolios`)
 * ------------------------------------------------------------------------- */

/**
 * Una fila del listado. `value` y `cost` van en número, no en texto, porque de
 * ellos sale el ancho de la barra de capital y ganancia.
 */
export interface TourPortfolio {
	name: string;
	/** El que se usa cuando no eliges otro. */
	isDefault?: boolean;
	detail: string;
	risk: string;
	value: number;
	cost: number;
	money: string;
	/** El mismo importe sin céntimos, para la columna estrecha del hero. */
	short: string;
	gain: string;
	up: boolean;
}

export const TOUR_PORTFOLIOS: TourPortfolio[] = [
	{
		name: 'Jubilación',
		isDefault: true,
		detail: 'Acciones y ETFs a largo plazo, 8 posiciones',
		risk: 'Moderado',
		value: 132400,
		cost: 116000,
		money: '$132,400.00',
		short: '$132,400',
		gain: '+14,14%',
		up: true
	},
	{
		name: 'Cripto',
		detail: 'Posición especulativa, revisada cada trimestre, 5 posiciones',
		risk: 'Agresivo',
		value: 68900,
		cost: 56600,
		money: '$68,900.00',
		short: '$68,900',
		gain: '+21,73%',
		up: true
	},
	{
		name: 'Reserva',
		detail: 'Colchón de liquidez y renta fija, 5 posiciones',
		risk: 'Conservador',
		value: 47200,
		cost: 48500,
		money: '$47,200.00',
		short: '$47,200',
		gain: '-2,68%',
		up: false
	}
];

/** El pie de la tabla: la misma cuenta del resumen, vista por abajo. */
export const TOUR_PORTFOLIO_TOTALS = {
	detail: '3 portafolios, 18 posiciones abiertas',
	cost: 'Capital invertido: $221,100.00',
	value: '$248,500.00',
	gain: '+12,40%'
};

/* ---------------------------------------------------------------------------
 * Transacciones (`/dashboard/transactions`)
 * ------------------------------------------------------------------------- */

export const TOUR_TRANSACTIONS = [
	{
		kind: 'Compra',
		note: 'Aporte mensual',
		asset: 'NVIDIA Corp.',
		ticker: 'NVDA',
		date: '24 de jun de 2026',
		qty: '20',
		price: '$121.80',
		total: '$2,436.00'
	},
	{
		kind: 'Dividendo',
		note: 'Dividendo trimestral',
		asset: 'Apple Inc.',
		ticker: 'AAPL',
		date: '18 de jun de 2026',
		qty: '42',
		price: '$0.26',
		total: '$10.92'
	},
	{
		kind: 'Compra',
		note: '',
		asset: 'Vanguard FTSE All-World UCITS ETF',
		ticker: 'VWCE',
		date: '11 de jun de 2026',
		qty: '15',
		price: '$126.40',
		total: '$1,896.00'
	},
	{
		kind: 'Interés',
		note: 'Intereses de la cuenta',
		asset: 'Efectivo en dólares',
		ticker: 'USD',
		date: '2 de jun de 2026',
		qty: '9.500',
		price: '$0.0021',
		total: '$19.95'
	},
	{
		kind: 'Venta',
		note: 'Toma de beneficios',
		asset: 'Ethereum',
		ticker: 'ETH',
		date: '27 de may de 2026',
		qty: '0,6',
		price: '$3,410.00',
		total: '$2,046.00'
	},
	{
		kind: 'Cargo',
		note: 'Custodia trimestral',
		asset: 'Efectivo en dólares',
		ticker: 'USD',
		date: '30 de abr de 2026',
		qty: '1',
		price: '$12.50',
		total: '$12.50'
	}
];

/* Los datos de la ficha de reportes viven en su propio módulo: eran la mitad de
   las cifras de este archivo. */
export {
	TOUR_MONTHS,
	TOUR_RECORD,
	TOUR_RETURNS,
	TOUR_KEY_STATS,
	tourReturnBackground
} from './tour-reports';
