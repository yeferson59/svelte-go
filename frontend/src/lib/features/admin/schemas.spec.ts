import { describe, it, expect } from 'vitest';
import {
	assetCreateSchema,
	assetPriceSchema,
	assetUpdateSchema,
	inviteUserSchema,
	rateCreateSchema,
	rateUpdateSchema,
	rowIdSchema
} from './schemas';

/** Primer mensaje de error, que es el que la action devuelve al formulario. */
function firstError(result: {
	success: boolean;
	error?: { issues: { message: string }[] };
}): string {
	return result.success ? '' : (result.error?.issues[0].message ?? '');
}

describe('inviteUserSchema', () => {
	it('normaliza el rol y recorta los espacios', () => {
		const parsed = inviteUserSchema.parse({
			email: '  nueva@finexia.test ',
			name: ' Nueva ',
			role: ' Admin '
		});

		expect(parsed).toEqual({ email: 'nueva@finexia.test', name: 'Nueva', role: 'admin' });
	});

	it('cae a customer cuando el rol llega vacío (atajo de la lista de espera)', () => {
		expect(inviteUserSchema.parse({ email: 'a@b.test', name: '', role: '' }).role).toBe('customer');
	});

	it('exige el correo antes que nada', () => {
		expect(firstError(inviteUserSchema.safeParse({ email: '  ', name: '', role: 'sin-rol' }))).toBe(
			'El correo es requerido'
		);
	});

	it('rechaza un rol que no existe', () => {
		expect(
			firstError(inviteUserSchema.safeParse({ email: 'a@b.test', name: '', role: 'root' }))
		).toBe('Rol inválido');
	});
});

describe('rowIdSchema', () => {
	it('acepta ids que no son UUID, como los de invitación', () => {
		expect(rowIdSchema.parse('invite-1')).toBe('invite-1');
	});

	it('rechaza un id vacío', () => {
		expect(firstError(rowIdSchema.safeParse('   '))).toBe('ID requerido');
	});
});

describe('assetCreateSchema', () => {
	it('normaliza ticker y moneda a mayúsculas', () => {
		const parsed = assetCreateSchema.parse({
			ticker: ' aapl ',
			name: ' Apple Inc. ',
			assetType: 'stock',
			currency: 'usd',
			exchange: ' NASDAQ '
		});

		expect(parsed).toEqual({
			ticker: 'AAPL',
			name: 'Apple Inc.',
			assetType: 'stock',
			currency: 'USD',
			exchange: 'NASDAQ'
		});
	});

	it('da el mismo aviso falte el campo que falte', () => {
		const message = 'Ticker, nombre, tipo y moneda son requeridos';
		for (const missing of ['ticker', 'name', 'assetType', 'currency']) {
			const input = { ticker: 'AAPL', name: 'Apple', assetType: 'stock', currency: 'USD' };
			const result = assetCreateSchema.safeParse({ ...input, [missing]: '' });
			expect(firstError(result)).toBe(message);
		}
	});

	it('deja el exchange vacío si no se indica', () => {
		const parsed = assetCreateSchema.parse({
			ticker: 'BTC',
			name: 'Bitcoin',
			assetType: 'crypto',
			currency: 'USD'
		});

		expect(parsed.exchange).toBe('');
	});
});

describe('assetPriceSchema', () => {
	it('acepta un precio positivo escrito como texto', () => {
		expect(assetPriceSchema.parse({ id: 'a1', price: '190.00', currency: 'USD' }).price).toBe(190);
	});

	it('rechaza precios no positivos o no numéricos', () => {
		expect(firstError(assetPriceSchema.safeParse({ id: 'a1', price: '0' }))).toBe(
			'Precio inválido'
		);
		expect(firstError(assetPriceSchema.safeParse({ id: 'a1', price: 'gratis' }))).not.toBe('');
	});

	it('exige el id del activo', () => {
		expect(firstError(assetPriceSchema.safeParse({ id: '', price: '10' }))).toBe(
			'ID de activo requerido'
		);
	});

	it('cae a USD cuando la fila no trae moneda', () => {
		expect(assetPriceSchema.parse({ id: 'a1', price: '10' }).currency).toBe('USD');
	});
});

describe('assetUpdateSchema', () => {
	/** Ficha completa tal como la manda el formulario de edición. */
	const full = {
		id: 'a1',
		ticker: ' aapl ',
		name: ' Apple Inc. ',
		assetType: 'stock',
		currency: 'usd',
		exchange: ' NASDAQ ',
		isCurated: 'on',
		price: ' 190.50 '
	};

	it('normaliza los mismos campos que el alta y conserva el id', () => {
		expect(assetUpdateSchema.parse(full)).toEqual({
			id: 'a1',
			ticker: 'AAPL',
			name: 'Apple Inc.',
			assetType: 'stock',
			currency: 'USD',
			exchange: 'NASDAQ',
			isCurated: true,
			price: '190.50'
		});
	});

	// El checkbox no manda nada cuando está desmarcado, así que `null` es la
	// forma en que un formulario dice «quítalo del catálogo compartido».
	it('lee el checkbox ausente como una despublicación explícita', () => {
		expect(assetUpdateSchema.parse({ ...full, isCurated: null }).isCurated).toBe(false);
		expect(assetUpdateSchema.parse({ ...full, isCurated: undefined }).isCurated).toBe(false);
	});

	it('admite el precio en blanco, que deja el guardado como está', () => {
		expect(assetUpdateSchema.parse({ ...full, price: '  ' }).price).toBe('');
	});

	it('conserva el precio como texto, con sus decimales de cola', () => {
		expect(assetUpdateSchema.parse({ ...full, price: '190.00' }).price).toBe('190.00');
	});

	it('rechaza un precio que no es un número positivo', () => {
		expect(firstError(assetUpdateSchema.safeParse({ ...full, price: '0' }))).toBe(
			'Precio inválido'
		);
		expect(firstError(assetUpdateSchema.safeParse({ ...full, price: 'gratis' }))).toBe(
			'Precio inválido'
		);
	});

	it('exige el id de la fila que se está editando', () => {
		expect(firstError(assetUpdateSchema.safeParse({ ...full, id: '  ' }))).toBe('ID requerido');
	});

	it('da el mismo aviso del alta falte el campo que falte', () => {
		for (const missing of ['ticker', 'name', 'assetType', 'currency']) {
			const result = assetUpdateSchema.safeParse({ ...full, [missing]: '' });
			expect(firstError(result)).toBe('Ticker, nombre, tipo y moneda son requeridos');
		}
	});
});

describe('rateCreateSchema', () => {
	it('normaliza las dos monedas a mayúsculas', () => {
		const parsed = rateCreateSchema.parse({
			fromCurrency: 'usd',
			toCurrency: ' cop ',
			rate: '4000'
		});

		expect(parsed).toEqual({ fromCurrency: 'USD', toCurrency: 'COP', rate: '4000' });
	});

	it('da el mismo aviso falte el campo que falte', () => {
		expect(
			firstError(rateCreateSchema.safeParse({ fromCurrency: 'USD', toCurrency: '', rate: '4000' }))
		).toBe('Moneda origen, destino y tasa son requeridos');
	});
});

describe('rateUpdateSchema', () => {
	it('rechaza una tasa no positiva', () => {
		expect(firstError(rateUpdateSchema.safeParse({ id: 'r1', rate: '-2' }))).toBe('Tasa inválida');
	});

	it('exige el id de la tasa', () => {
		expect(firstError(rateUpdateSchema.safeParse({ id: '', rate: '4000' }))).toBe(
			'ID de tasa requerido'
		);
	});
});
