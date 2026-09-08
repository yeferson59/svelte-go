/**
 * Fixtures del stub de la API.
 *
 * Vivían dentro de `mock-api.mjs`, mezcladas con el enrutado. Se separan porque
 * han dejado de ser cuatro constantes: describen una cuenta completa —tres
 * portafolios, veinte posiciones, catorce meses de historial— y de ellas salen
 * tanto los e2e como las capturas del manual de usuario
 * (`pnpm manual:shots`). Con un solo portafolio y dos movimientos, el manual
 * enseñaba gráficas vacías y una proyección que la aplicación se negaba a
 * calcular por falta de historial.
 *
 * Todo lo que se puede derivar, se deriva: los totales de cada portafolio salen
 * de sus posiciones y la serie de crecimiento termina exactamente en el valor
 * que suman. Así las cifras cuadran entre pantallas y nadie tiene que
 * recalcular a mano una fixture cuando toca otra.
 */

export const NOW = '2026-07-01T00:00:00Z';
export const FUTURE = '2027-01-01T00:00:00Z';

export const IDS = {
	portfolio: '11111111-1111-4111-8111-111111111111',
	portfolioCrypto: '11111111-1111-4111-8111-111111111112',
	portfolioReserve: '11111111-1111-4111-8111-111111111113',
	platform: '33333333-3333-4333-8333-333333333333',
	platformExchange: '33333333-3333-4333-8333-333333333334',
	platformBank: '33333333-3333-4333-8333-333333333335',
	riskModerate: '44444444-4444-4444-8444-444444444444',
	riskAggressive: '44444444-4444-4444-8444-444444444445',
	riskConservative: '44444444-4444-4444-8444-444444444446',
	entry: '55555555-5555-4555-8555-555555555555'
};

const money = (value) => value.toFixed(2);

// --- Catálogo de activos ----------------------------------------------------

/*
 * `[ticker, nombre, tipo, mercado, precio de mercado]`.
 *
 * No hay columna de «categoría»: la migración 000026 borró
 * `portfolio_entries.category` —una copia del tipo del activo que se quedaba
 * obsoleta al reclasificarlo— junto con su enum en plural. Hoy el backend hace
 * `entry.Category = entry.Asset.AssetType`, así que la categoría de una
 * posición *es* su tipo, en singular.
 */
const CATALOG = [
	['AAPL', 'Apple Inc.', 'stock', 'NASDAQ', 214.35],
	['MSFT', 'Microsoft Corp.', 'stock', 'NASDAQ', 438.9],
	['NVDA', 'NVIDIA Corp.', 'stock', 'NASDAQ', 126.4],
	['VWCE', 'Vanguard FTSE All-World UCITS ETF', 'etf', 'XETRA', 128.62],
	['CSPX', 'iShares Core S&P 500 UCITS ETF', 'etf', 'XETRA', 568.2],
	['BTC', 'Bitcoin', 'crypto', '', 67240.0],
	['ETH', 'Ethereum', 'crypto', '', 3482.5],
	['SOL', 'Solana', 'crypto', '', 168.9],
	['TLT', 'iShares 20+ Year Treasury Bond ETF', 'bond', 'NASDAQ', 92.15],
	['USD', 'Efectivo en dólares', 'cash', '', 1.0]
];

const asset = (ticker) => CATALOG.find((row) => row[0] === ticker);

/** `GET /portfolios/assets` — el catálogo que ve el buscador de activos. */
export const assets = CATALOG.map(([ticker, name, assetType, exchange, price], index) => ({
	id: `22222222-2222-4222-8222-2222222222${String(index).padStart(2, '0')}`,
	ticker,
	name,
	assetType,
	exchange,
	currency: 'USD',
	currentPrice: { value: money(price), currency: 'USD' },
	// BTC lo aportó la usuaria y no tiene precio manual fechado del catálogo.
	priceUpdatedAt: ticker === 'BTC' ? null : NOW,
	isCurated: ticker !== 'BTC'
}));

const assetId = (ticker) => assets.find((a) => a.ticker === ticker).id;

// --- Posiciones -------------------------------------------------------------

let entrySeq = 0;

/** Posición dentro de un portafolio: cantidad y precio medio de compra. */
function holding(ticker, quantity, price, entryDate, notes = '') {
	const [, name, assetType, exchange, marketPrice] = asset(ticker);
	entrySeq += 1;
	return {
		// La primera posición conserva el id histórico: los e2e lo usan.
		id:
			entrySeq === 1
				? IDS.entry
				: `55555555-5555-4555-8555-5555555555${String(entrySeq).padStart(2, '0')}`,
		assetId: assetId(ticker),
		ticker,
		name,
		assetType,
		exchange,
		currency: 'USD',
		quantity: String(quantity),
		price: money(price),
		marketPrice: money(marketPrice),
		costCurrency: 'USD',
		// Totales en la moneda base del portafolio. Aquí todo está en USD, así
		// que coinciden con cantidad × precio; el backend los envía siempre.
		costBasisBase: money(quantity * price),
		marketValueBase: money(quantity * marketPrice),
		fxConverted: true,
		// El backend la copia del tipo del activo; no es un dato aparte.
		category: assetType,
		entryDate,
		notes
	};
}

export const PORTFOLIOS = [
	{
		id: IDS.portfolio,
		name: 'Cartera Principal',
		description: 'Acciones y ETFs a largo plazo',
		type: 'stocks_etfs',
		riskId: IDS.riskModerate,
		riskName: 'Moderado',
		isDefault: true,
		holdings: [
			holding('AAPL', 42, 168.4, '2025-06-12', 'Núcleo de la cartera'),
			holding('MSFT', 18, 372.5, '2025-07-03'),
			holding('VWCE', 120, 108.75, '2025-06-12', 'Aporte mensual'),
			holding('CSPX', 9, 492.3, '2025-09-18'),
			holding('NVDA', 60, 98.2, '2026-01-15')
		]
	},
	{
		id: IDS.portfolioCrypto,
		name: 'Cripto',
		description: 'Posición especulativa, revisada cada trimestre',
		type: 'cryptos',
		riskId: IDS.riskAggressive,
		riskName: 'Agresivo',
		isDefault: false,
		holdings: [
			holding('BTC', 0.15, 54300.0, '2025-08-04'),
			holding('ETH', 2.2, 2980.0, '2025-08-04'),
			holding('SOL', 25, 142.6, '2026-02-20')
		]
	},
	{
		id: IDS.portfolioReserve,
		name: 'Reserva',
		description: 'Colchón de liquidez y renta fija',
		type: 'bonds',
		riskId: IDS.riskConservative,
		riskName: 'Conservador',
		isDefault: false,
		holdings: [holding('TLT', 140, 96.4, '2025-10-09'), holding('USD', 9500, 1.0, '2025-06-01')]
	}
];

/** Todas las posiciones, sin importar el portafolio. */
export const holdings = PORTFOLIOS.flatMap((p) => p.holdings);

const costOf = (h) => Number(h.quantity) * Number(h.price);
const valueOf = (h) => Number(h.quantity) * Number(h.marketPrice);

const sum = (values) => values.reduce((acc, v) => acc + v, 0);

export const TOTAL_COST = sum(holdings.map(costOf));
export const TOTAL_VALUE = sum(holdings.map(valueOf));

/** `GET /portfolios/summary` — totales derivados de las posiciones. */
export const portfolioSummary = (displayCurrency = 'USD') =>
	PORTFOLIOS.map((p) => {
		const cost = sum(p.holdings.map(costOf));
		const value = sum(p.holdings.map(valueOf));
		const gain = value - cost;
		return {
			id: p.id,
			name: p.name,
			description: p.description,
			type: p.type,
			baseCurrency: 'USD',
			displayCurrency,
			// El backend marca cada resumen con si pudo convertirlo: lo que no
			// puede se queda en su moneda y no entra en las sumas de la página.
			fxConverted: true,
			isDefault: p.isDefault,
			riskId: p.riskId,
			riskName: p.riskName,
			totalPositions: p.holdings.length,
			totalCostBase: money(cost),
			totalMarketValue: money(value),
			totalGainLoss: money(gain),
			totalGainLossPct: (cost > 0 ? (gain / cost) * 100 : 0).toFixed(2),
			// Todas las posiciones del fixture están en USD, la moneda base, así
			// que no hay ninguna que convertir ni de la que avisar.
			positionsUnconverted: 0,
			createdAt: NOW
		};
	});

/*
 * `GET /portfolios/allocation` — reparto por clase de activo, en el orden del
 * donut. Agrupa por `assetType` porque es por lo que agrupa el backend desde
 * que lee `assets.asset_type`: las claves son `stock`, `etf`… y es el
 * vocabulario que el donut traduce a etiquetas y colores.
 */
export const allocation = (() => {
	const byCategory = new Map();
	for (const h of holdings) {
		byCategory.set(h.assetType, (byCategory.get(h.assetType) ?? 0) + valueOf(h));
	}
	return [...byCategory.entries()]
		.sort(([, a], [, b]) => b - a)
		.map(([category, value]) => ({
			category,
			marketValue: money(value),
			percent: Number(((value / TOTAL_VALUE) * 100).toFixed(2)),
			// El backend responde el reparto en una sola moneda y dice cuál es;
			// aquí todo está en USD, así que no hay nada sin convertir.
			currency: 'USD',
			positionsUnconverted: 0
		}));
})();

/*
 * `GET /portfolios/holdings` — una fila por activo, sumando todos los
 * portafolios. Es el mismo dinero que reparte `allocation`, un nivel más fino:
 * por activo en vez de por clase, y con las unidades, que solo esta vista trae.
 *
 * Se deriva de las posiciones igual que todo lo demás, así que el total de la
 * lista es el mismo `TOTAL_VALUE` que enseña el resto de las pantallas.
 */
export const assetHoldings = (() => {
	const byTicker = new Map();

	for (const portfolio of PORTFOLIOS) {
		for (const h of portfolio.holdings) {
			const row = byTicker.get(h.ticker) ?? {
				position: h,
				quantity: 0,
				value: 0,
				portfolios: new Set()
			};
			row.quantity += Number(h.quantity);
			row.value += valueOf(h);
			row.portfolios.add(portfolio.id);
			byTicker.set(h.ticker, row);
		}
	}

	return [...byTicker.values()]
		.sort((a, b) => b.value - a.value)
		.map(({ position, quantity, value, portfolios }) => ({
			assetId: position.assetId,
			ticker: position.ticker,
			name: position.name,
			assetType: position.assetType,
			exchange: position.exchange,
			currency: 'USD',
			quantity: String(quantity),
			marketPrice: position.marketPrice,
			marketValue: money(value),
			percent: Number(((value / TOTAL_VALUE) * 100).toFixed(2)),
			displayCurrency: 'USD',
			portfolios: portfolios.size,
			// Todo el catálogo del fixture tiene precio y está en USD: nada se
			// valora a coste y no hay nada que convertir.
			priceSource: 'own',
			// Con `priceSource: 'own'` el backend dice de qué clave salió el
			// precio y cuándo: es lo que el panel del activo necesita para saber
			// a quién repreguntar.
			priceProvider: 'finnhub',
			priceFetchedAt: NOW,
			positionsUnconverted: 0
		}));
})();

// --- Serie de crecimiento ---------------------------------------------------

/*
 * Catorce meses, cinco puntos por mes. Las curvas son fracciones del valor y
 * del coste finales, así que la serie termina exactamente en lo que suman las
 * posiciones. El último punto de cada mes cae justo en su ancla: el calendario
 * de rentabilidad de reportes lee ese punto, y así sus porcentajes son los de
 * las curvas y no un artefacto de la interpolación.
 */
const VALUE_CURVE = [
	0.78, 0.8, 0.83, 0.81, 0.85, 0.88, 0.86, 0.9, 0.93, 0.95, 0.93, 0.96, 0.98, 1.0
];
const COST_CURVE = [
	0.74, 0.76, 0.79, 0.79, 0.82, 0.85, 0.85, 0.88, 0.91, 0.93, 0.93, 0.96, 0.98, 1.0
];

/** Ondulación fija dentro del mes: la serie tiene que ser reproducible. */
const WOBBLE = [0.004, -0.003, 0.006, -0.005, 0.002, -0.004, 0.005, -0.002];

const FIRST_MONTH = { year: 2025, month: 5 }; // junio de 2025 (0-based)
const DAYS = [1, 8, 15, 22, 28];

function isoDate(year, month, day) {
	const date = new Date(Date.UTC(year, month, day));
	return date.toISOString().slice(0, 10);
}

export const growth = (() => {
	const points = [];
	let wobbleIndex = 0;

	let prevPointCost = null;

	for (let m = 0; m < VALUE_CURVE.length; m++) {
		const prevValue = m === 0 ? VALUE_CURVE[0] * 0.98 : VALUE_CURVE[m - 1];
		const prevCost = m === 0 ? COST_CURVE[0] * 0.99 : COST_CURVE[m - 1];

		for (const day of DAYS) {
			const t = day / 28;
			const last = day === 28;
			const noise = last ? 0 : WOBBLE[wobbleIndex++ % WOBBLE.length];

			const value = TOTAL_VALUE * (prevValue + (VALUE_CURVE[m] - prevValue) * t + noise);
			const cost = TOTAL_COST * (prevCost + (COST_CURVE[m] - prevCost) * t);
			const gain = value - cost;
			// La serie del stub solo compra, y en una compra el aporte es
			// exactamente lo que sube el capital invertido: el backend lo saca
			// de las transacciones, aquí sale de la curva que las representa.
			const netFlow = prevPointCost === null ? 0 : cost - prevPointCost;
			prevPointCost = cost;

			points.push({
				date: isoDate(FIRST_MONTH.year, FIRST_MONTH.month + m, day),
				totalValue: money(value),
				totalCostBase: money(cost),
				gainLoss: money(gain),
				gainLossPct: (cost > 0 ? (gain / cost) * 100 : 0).toFixed(2),
				netFlow: money(netFlow)
			});
		}
	}

	const first = points[0];
	const last = points[points.length - 1];
	const initial = Number(first.totalValue);
	const current = Number(last.totalValue);

	return {
		points,
		summary: {
			firstDate: first.date,
			initialValue: first.totalValue,
			currentValue: last.totalValue,
			totalGrowthPct: (((current - initial) / initial) * 100).toFixed(2)
		}
	};
})();

/** Serie de un portafolio: la agregada, a escala de lo que pesa dentro. */
export function growthFor(portfolioId) {
	const summary = portfolioSummary().find((p) => p.id === portfolioId);
	if (!summary) return growth;

	const share = Number(summary.totalMarketValue) / TOTAL_VALUE;
	const points = growth.points.map((point) => {
		const value = Number(point.totalValue) * share;
		const cost = Number(point.totalCostBase) * share;
		return {
			date: point.date,
			totalValue: money(value),
			totalCostBase: money(cost),
			gainLoss: money(value - cost),
			gainLossPct: (cost > 0 ? ((value - cost) / cost) * 100 : 0).toFixed(2),
			netFlow: money(Number(point.netFlow) * share)
		};
	});

	return {
		points,
		summary: {
			firstDate: points[0].date,
			initialValue: points[0].totalValue,
			currentValue: points[points.length - 1].totalValue,
			totalGrowthPct: growth.summary.totalGrowthPct
		}
	};
}

// --- Movimientos ------------------------------------------------------------

const entryIdOf = (ticker) => holdings.find((h) => h.ticker === ticker).id;

/** `[fecha, tipo, ticker, cantidad, precio, comisión, nota]`. */
const MOVEMENTS = [
	['2026-06-24', 'buy', 'NVDA', '20', '121.80', '1.20', 'Aporte mensual'],
	['2026-06-18', 'dividend', 'AAPL', '42', '0.26', '0.00', 'Dividendo trimestral'],
	['2026-06-11', 'buy', 'VWCE', '15', '126.40', '1.00', ''],
	['2026-06-02', 'interest', 'USD', '9500', '0.0021', '0.00', 'Intereses de la cuenta'],
	['2026-05-27', 'sell', 'ETH', '0.6', '3410.00', '4.10', 'Toma de beneficios'],
	['2026-05-19', 'buy', 'BTC', '0.03', '61200.00', '9.80', ''],
	['2026-05-08', 'dividend', 'TLT', '140', '0.31', '0.00', 'Cupón mensual'],
	['2026-04-30', 'fee', 'USD', '1', '12.50', '0.00', 'Custodia trimestral'],
	['2026-04-21', 'buy', 'CSPX', '2', '541.60', '1.50', ''],
	['2026-04-09', 'transfer_in', 'USD', '2500', '1.00', '0.00', 'Traspaso desde el banco'],
	['2026-03-17', 'buy', 'SOL', '10', '151.20', '2.40', ''],
	['2026-02-20', 'buy', 'AAPL', '12', '186.90', '1.20', '']
];

export const transactions = MOVEMENTS.map(
	([date, type, ticker, quantity, price, fees, notes], index) => ({
		// El backend identifica las transacciones con UUID, y las acciones que
		// las editan o borran lo validan como tal: un `txn-1` no llegaría nunca
		// a la API.
		id: `66666666-6666-4666-8666-6666666666${String(index + 1).padStart(2, '0')}`,
		entryId: entryIdOf(ticker),
		type,
		quantity,
		price,
		currency: 'USD',
		fees,
		transactionDate: `${date}T00:00:00Z`,
		notes,
		createdAt: `${date}T00:00:00Z`,
		assetTicker: ticker,
		assetName: asset(ticker)[1]
	})
);

/** Mayor movimiento de un portafolio, por importe. */
export function topTransaction(portfolioId) {
	const tickers = new Set(
		(PORTFOLIOS.find((p) => p.id === portfolioId)?.holdings ?? []).map((h) => h.ticker)
	);
	const candidates = transactions.filter((t) => tickers.has(t.assetTicker));
	if (candidates.length === 0) return null;

	const top = candidates.reduce((best, t) =>
		Number(t.quantity) * Number(t.price) > Number(best.quantity) * Number(best.price) ? t : best
	);

	return {
		value: money(Number(top.quantity) * Number(top.price)),
		type: top.type,
		currency: 'USD',
		assetTicker: top.assetTicker,
		assetName: top.assetName,
		transactionDate: top.transactionDate
	};
}

// --- Plataformas y tasas ----------------------------------------------------

/**
 * Valor custodiado en cada plataforma, repartiendo las posiciones por tipo.
 *
 * `totalValue` es lo **invertido** y `marketValue` lo que vale hoy, como en el
 * backend: si el stub mandara el valor de mercado en los dos, la vista pintaría
 * una ganancia de cero y la pantalla que la explica no se podría probar.
 */
const platformFigures = (holdings) => {
	const cost = sum(holdings.map(costOf));
	const value = sum(holdings.map(valueOf));

	return {
		investments: holdings.length,
		assets: holdings.length,
		// Cada plataforma del stub custodia las posiciones de un solo portafolio.
		portfolios: 1,
		totalValue: money(cost),
		marketValue: money(value),
		gainLoss: money(value - cost),
		gainLossPct: Number((((value - cost) / cost) * 100).toFixed(2)),
		displayCurrency: 'USD',
		positionsPricedOwn: holdings.length,
		positionsPricedManual: 0,
		positionsAtCost: 0,
		positionsUnconverted: 0
	};
};

const PLATFORM_COST = sum(PORTFOLIOS.map((p) => sum(p.holdings.map(costOf))));

export const sources = [
	{
		id: IDS.platform,
		name: 'Broker Demo',
		description: 'Acciones y ETFs',
		sourceType: 'broker',
		isActive: true,
		...platformFigures(PORTFOLIOS[0].holdings),
		createdAt: NOW
	},
	{
		id: IDS.platformExchange,
		name: 'Exchange Demo',
		description: 'Criptomonedas',
		sourceType: 'crypto_wallet',
		isActive: true,
		...platformFigures(PORTFOLIOS[1].holdings),
		createdAt: NOW
	},
	{
		id: IDS.platformBank,
		name: 'Banco Demo',
		description: 'Renta fija y efectivo',
		sourceType: 'investment_bank',
		isActive: true,
		...platformFigures(PORTFOLIOS[2].holdings),
		createdAt: NOW
	}
	// `percent` necesita el conjunto, así que se reparte una vez construidas.
].map((source) => ({
	...source,
	percent: Number(((Number(source.totalValue) / PLATFORM_COST) * 100).toFixed(2))
}));

// `source` reproduce el reparto real: el feed público solo publica USD/COP (la
// TRM), y todo lo demás lo escribió un administrador.
export const exchangeRates = [
	{
		id: '66666666-6666-4666-8666-666666666661',
		fromCurrency: 'USD',
		toCurrency: 'COP',
		rate: '4123.456789',
		rateDate: NOW,
		source: 'dolarapi',
		createdAt: NOW
	},
	{
		id: '66666666-6666-4666-8666-666666666662',
		fromCurrency: 'EUR',
		toCurrency: 'USD',
		rate: '1.085',
		rateDate: NOW,
		source: 'manual',
		createdAt: NOW
	},
	{
		id: '66666666-6666-4666-8666-666666666663',
		fromCurrency: 'GBP',
		toCurrency: 'USD',
		rate: '1.272',
		rateDate: NOW,
		source: 'manual',
		createdAt: NOW
	}
];

export const risks = [
	{ id: IDS.riskConservative, name: 'Conservador', description: 'Prioriza preservar el capital' },
	{
		id: IDS.riskModerate,
		name: 'Moderado',
		description: 'Equilibrio entre crecimiento y estabilidad'
	},
	{ id: IDS.riskAggressive, name: 'Agresivo', description: 'Busca máximo crecimiento' }
];

// --- Importación ------------------------------------------------------------

/*
 * Vista previa con filas descartadas a propósito: el manual explica que la
 * pantalla avisa de las que se omitirán y por qué, y con una previsualización
 * impecable esa parte de la interfaz no se veía nunca.
 */
const IMPORT_ROWS = [
	['2026-06-24', 'buy', 'NVDA', '20', '121.80', true, []],
	['2026-06-18', 'dividend', 'AAPL', '42', '0.26', true, []],
	['2026-06-11', 'buy', 'VWCE', '15', '126.40', true, []],
	['24/06/26', 'buy', 'MSFT', '3', '431.00', false, ['Fecha no reconocida']],
	['2026-05-27', 'sell', 'ETH', '1.5', '3410.00', true, []],
	['2026-05-19', 'buy', 'BTC', '0.06', '61200.00', true, []],
	['2026-05-08', 'buy', 'CSPX', '2', '', false, ['Precio vacío']],
	['2026-04-21', 'buy', 'TLT', '40', '95.10', true, []]
];

export const importPreview = {
	sheets: ['Movimientos', 'Resumen'],
	sheet: 'Movimientos',
	headerRow: 1,
	headers: ['Fecha', 'Tipo', 'Ticker', 'Cantidad', 'Precio'],
	suggestedMapping: {
		date: 0,
		type: 1,
		ticker: 2,
		assetName: null,
		quantity: 3,
		price: 4,
		fees: null,
		currency: null,
		category: null,
		notes: null
	},
	missingFields: [],
	totalRows: IMPORT_ROWS.length,
	validRows: IMPORT_ROWS.filter(([, , , , , valid]) => valid).length,
	invalidRows: IMPORT_ROWS.filter(([, , , , , valid]) => !valid).length,
	rows: IMPORT_ROWS.map(([date, type, ticker, quantity, price, valid, errors], index) => ({
		rowNumber: index + 2,
		raw: [date, type, ticker, quantity, price],
		date: valid ? date : '',
		type,
		ticker,
		assetName: asset(ticker)?.[1] ?? ticker,
		quantity,
		price,
		fees: '',
		currency: 'USD',
		category: asset(ticker)?.[2] ?? 'other',
		notes: '',
		valid,
		errors
	}))
};

/*
 * `GET /market/credentials` — las claves de proveedor de la cuenta.
 *
 * Nunca traen la clave: proveedor, cuatro últimos caracteres y estado, que es
 * todo lo que el backend sirve de ellas. Dos filas y no una porque el panel de
 * un activo se prueba eligiendo entre proveedores, y con una sola no hay nada
 * que elegir.
 */
export const marketCredentials = [
	{
		provider: 'finnhub',
		last4: '3f9a',
		status: 'active',
		lastVerifiedAt: NOW,
		createdAt: NOW,
		updatedAt: NOW
	},
	{
		provider: 'alphavantage',
		last4: '1c2d',
		status: 'active',
		lastVerifiedAt: NOW,
		createdAt: NOW,
		updatedAt: NOW
	}
];
