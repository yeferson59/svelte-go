/**
 * Vista consolidada de activos: lo que el usuario tiene de cada uno sumando
 * todos sus portafolios (`GET /portfolios/holdings`).
 *
 * Responde a dos preguntas que ninguna otra pantalla contesta: «¿cuánto tengo
 * de X?» —las posiciones del detalle solo suman dentro de su portafolio, y el
 * reparto del panel pliega todo a ocho clases de activo, así que un activo
 * repartido entre tres portafolios no tenía una sola fila en ninguna parte— y
 * «¿está mi dinero en pocas manos?», que es lo que mide la barra de
 * concentración.
 *
 * Helpers puros: sin Svelte y sin red. Los contratos vienen de `$lib/api/types`.
 */
import type { AssetHolding } from '$lib/api/types';
import { formatAssetType } from '$lib/shared/format/asset-type';

export type { AssetHolding };

/**
 * Cuántos activos se dibujan con su propio color antes de agrupar la cola.
 *
 * No está para que la barra «se lea» —una barra apilada aguanta muchas más
 * franjas que una torta— sino para que la cola no se coma el dibujo: una
 * cartera de doscientas posiciones son ciento ochenta franjas de menos de un
 * píxel, y lo que se ve entonces es la separación entre ellas, no los datos.
 */
export const BAND_MAX = 12;

/** Gris cálido de la cola agrupada: no es un puesto más, es lo que queda. */
export const REST_COLOR = '#4a4642';

/**
 * Color de una franja según su puesto en el ranking.
 *
 * Un solo tono —el ámbar con el que toda la aplicación pinta el valor de
 * mercado— aclarándose hacia el activo mayor. El puesto es un orden, no una
 * categoría: la luminosidad es el canal que dice «orden», y el matiz el que
 * dice «grupo». Los siete matices que había antes afirmaban siete grupos de un
 * dato que solo tiene un eje, y encima no sobrevivían a la comprobación de
 * daltonismo: dos acciones distintas salían del mismo color, y el azul y el
 * morado quedaban a una diferencia que en protanopia no existe. Con un matiz
 * único no hay pareja que confundir.
 *
 * En OKLCH y no en HSL porque la luminosidad de OKLab sí es perceptual: los
 * pasos intermedios del ámbar en HSL se apelotonan y los oscuros se separan de
 * más, que es justo lo contrario de lo que necesita una escala de puestos.
 */
export function rankColor(index: number, count: number): string {
	const t = count <= 1 ? 0 : Math.min(index, count - 1) / (count - 1);
	const lightness = 0.78 - t * 0.34;
	// La croma baja con la luminosidad: mantenerla arriba saca del gamut los
	// pasos oscuros y el navegador los recorta todos al mismo naranja plano.
	const chroma = 0.13 - t * 0.055;

	return `oklch(${lightness.toFixed(3)} ${chroma.toFixed(3)} 71)`;
}

/** Fila de la tabla: el holding del backend ya convertido a números. */
export interface AssetHoldingRow {
	assetId: string;
	ticker: string;
	name: string;
	assetType: string;
	/** Etiqueta legible de la clase de activo. */
	typeLabel: string;
	/** Unidades, sumadas entre portafolios. Solo significan algo en su fila. */
	quantity: number;
	/** Precio por unidad en `currency`, o `null` si la posición va a coste. */
	marketPrice: number | null;
	/** Moneda en la que cotiza el activo (la de `marketPrice`). */
	currency: string;
	/** Valor de la posición, en la moneda de visualización. */
	value: number;
	percent: number;
	portfolios: number;
	priceSource: string;
	/**
	 * De qué proveedor salió `marketPrice` y cuándo, o `null` cuando no salió de
	 * ninguno (precio manual del catálogo, o posición valorada a coste).
	 *
	 * `priceSource` dice que el precio es del propio usuario, que basta para
	 * saber que no es un coste y no basta para decidir nada: volver a preguntar
	 * —y a quién— necesita el nombre y la hora.
	 */
	priceProvider: string | null;
	priceFetchedAt: string | null;
	/**
	 * `false` si alguna posición del activo no pudo convertirse por falta de
	 * tasa: su importe va en su moneda nativa y el total la mezcla.
	 */
	fxConverted: boolean;
}

/**
 * Precio por unidad, o `null` cuando no hay ninguno.
 *
 * El backend manda cadena vacía para una posición valorada a coste: cada
 * entrada pagó lo suyo y ningún número representa al activo. Vacío no es cero
 * —un 0 se leería como un activo que no vale nada— y por eso no se parsea.
 */
function toPrice(raw: string): number | null {
	if (raw === '') return null;
	const parsed = parseFloat(raw);

	return Number.isFinite(parsed) ? parsed : null;
}

export function toAssetHoldingRows(holdings: AssetHolding[]): AssetHoldingRow[] {
	return holdings.map((h) => ({
		assetId: h.assetId,
		ticker: h.ticker,
		name: h.name,
		assetType: h.assetType,
		typeLabel: formatAssetType(h.assetType),
		quantity: parseFloat(h.quantity) || 0,
		// Vacío no es cero: es «no hay precio que represente al activo», y un 0
		// se leería como un activo que no vale nada.
		marketPrice: toPrice(h.marketPrice),
		currency: h.currency,
		value: parseFloat(h.marketValue) || 0,
		percent: h.percent,
		portfolios: h.portfolios,
		priceSource: h.priceSource,
		// Ausentes salvo cuando el precio lo trajo una clave del usuario; se
		// normalizan a `null` para que la plantilla pregunte una sola cosa.
		priceProvider: h.priceProvider || null,
		priceFetchedAt: h.priceFetchedAt || null,
		fxConverted: h.positionsUnconverted === 0
	}));
}

/** Franja de la barra de concentración: un activo, o la cola agrupada. */
export interface BandSegment {
	/** Clave de la franja: el ticker, o `__rest__` para la cola. */
	key: string;
	label: string;
	value: number;
	/**
	 * Ancho de la franja, en porcentaje de la barra.
	 *
	 * No es `percent`. Los anchos tienen que sumar 100 o la barra deja un hueco
	 * al final, y los pesos del backend no suman 100 exacto: están redondeados a
	 * dos decimales y se calculan sobre un total que incluye posiciones que aquí
	 * no se dibujan. Así que el ancho se reparte aquí y el peso se imprime tal
	 * como vino.
	 */
	width: number;
	/** Peso sobre el total, el que calculó el backend. Es el que se imprime. */
	percent: number;
	color: string;
	/** Cuántos activos hay detrás: 1, salvo en la cola. */
	assets: number;
}

/**
 * Reparte las filas en franjas: los `max` mayores con su color y el resto
 * agrupado al final.
 *
 * Ordenadas de mayor a menor, que es lo que convierte la barra en una lectura:
 * la mitad izquierda es la mitad del dinero, y lo que se mira es cuántas
 * franjas caben en ella.
 */
export function buildBand(rows: AssetHoldingRow[], max = BAND_MAX): BandSegment[] {
	const ranked = [...rows].sort((a, b) => b.value - a.value).filter((row) => row.value > 0);
	const total = ranked.reduce((sum, row) => sum + row.value, 0);

	if (total <= 0) return [];

	// Un solo activo de más no merece una franja «resto» que lo esconda y ocupe
	// lo mismo: con `max + 1` caben todos.
	const cut = ranked.length <= max + 1 ? ranked.length : max;
	const head = ranked.slice(0, cut);
	const tail = ranked.slice(cut);
	const count = head.length + (tail.length > 0 ? 1 : 0);

	const segments: BandSegment[] = head.map((row, i) => ({
		key: row.ticker,
		label: row.ticker,
		value: row.value,
		width: (row.value / total) * 100,
		percent: row.percent,
		color: rankColor(i, count),
		assets: 1
	}));

	if (tail.length > 0) {
		const value = tail.reduce((sum, row) => sum + row.value, 0);
		segments.push({
			key: '__rest__',
			label: 'Resto',
			value,
			width: (value / total) * 100,
			percent: tail.reduce((sum, row) => sum + row.percent, 0),
			color: REST_COLOR,
			assets: tail.length
		});
	}

	return segments;
}

/**
 * Cuántos de los mayores activos hacen falta para llegar a la mitad del valor.
 *
 * Es la lectura de la barra puesta en número: la marca del centro cae en la
 * mitad del dinero —los anchos son proporcionales al valor, así que el punto
 * medio lo es por construcción— y esto dice cuántas franjas quedan a su
 * izquierda. Dos, y la cartera está concentrada; quince, y está repartida.
 *
 * Cuenta sobre el ranking entero, no sobre las franjas dibujadas: si la mitad
 * del valor no se alcanza hasta el activo dieciocho, eso es exactamente lo que
 * hay que poder decir.
 */
export function halfValueCount(rows: AssetHoldingRow[]): number {
	const ranked = [...rows].sort((a, b) => b.value - a.value).filter((row) => row.value > 0);
	const half = ranked.reduce((sum, row) => sum + row.value, 0) / 2;

	if (half <= 0) return 0;

	let running = 0;
	for (let i = 0; i < ranked.length; i++) {
		running += ranked[i].value;
		if (running >= half) return i + 1;
	}

	return ranked.length;
}

/**
 * Unidades de un activo, con los decimales que hagan falta.
 *
 * No es dinero y no lleva su formato: 0,00000123 BTC y 15 acciones son la misma
 * columna, y redondear a dos decimales convierte la primera en cero. Ocho es lo
 * que guarda la base de datos.
 */
export function formatQuantity(value: number): string {
	return new Intl.NumberFormat('es-CO', { maximumFractionDigits: 8 }).format(value);
}
