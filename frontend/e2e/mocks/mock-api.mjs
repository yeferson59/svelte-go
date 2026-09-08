// Backend stub for the e2e suite. Serves the subset of the HTTP contract
// (docs/API.md) that the SvelteKit server consumes, with fixed fixtures, so
// the smoke tests exercise the frontend's loaders/actions/session handling
// without a real Go backend or database. Playwright starts it via the
// `webServer` array in playwright.config.ts and points BASE_API at it.
import { createServer } from 'node:http';
import { pathToFileURL } from 'node:url';

const PORT = Number(process.env.MOCK_API_PORT ?? 4174);
const API_PREFIX = '/api/v1';

export const PASSWORD = 'Password123!';

// Las fixtures viven aparte: describen una cuenta completa y las usan tanto
// esta suite como el generador de capturas del manual (`pnpm manual:shots`).
import {
	FUTURE,
	IDS,
	NOW,
	allocation,
	assetHoldings,
	assets,
	exchangeRates,
	growth,
	growthFor,
	holdings,
	importPreview,
	marketCredentials,
	PORTFOLIOS,
	portfolioSummary,
	risks,
	sources,
	topTransaction,
	transactions
} from './fixtures.mjs';

// Se reexportan para `contract.spec.ts`, que valida las fixtures contra los
// schemas Zod de los que salen los tipos de la aplicación.
export {
	assets,
	exchangeRates,
	growth,
	holdings,
	portfolioSummary,
	sources,
	transactions,
	allocation,
	assetHoldings,
	marketCredentials
};

const ACCOUNTS = {
	'user@finexia.test': {
		accessToken: 'access-user',
		refreshToken: 'refresh-user',
		user: {
			name: 'Usuaria Prueba',
			email: 'user@finexia.test',
			emailVerified: true,
			image: '',
			role: 'customer',
			preferredCurrency: 'USD',
			createdAt: NOW,
			updatedAt: NOW
		},
		session: {
			id: 'session-user',
			userId: 'user-1',
			expiresAt: FUTURE,
			ipAddress: null,
			userAgent: null,
			createdAt: NOW
		}
	},
	// Cuenta sin verificar: la página de notificaciones avisa de que, mientras
	// la dirección no esté confirmada, no puede salir ningún correo.
	'sinverificar@finexia.test': {
		accessToken: 'access-unverified',
		refreshToken: 'refresh-unverified',
		user: {
			name: 'Usuario Sin Verificar',
			email: 'sinverificar@finexia.test',
			emailVerified: false,
			image: '',
			role: 'customer',
			preferredCurrency: 'USD',
			createdAt: NOW,
			updatedAt: NOW
		},
		session: {
			id: 'session-unverified',
			userId: 'user-2',
			expiresAt: FUTURE,
			ipAddress: null,
			userAgent: null,
			createdAt: NOW
		}
	},
	'admin@finexia.test': {
		accessToken: 'access-admin',
		refreshToken: 'refresh-admin',
		user: {
			name: 'Admin Prueba',
			email: 'admin@finexia.test',
			emailVerified: true,
			image: '',
			role: 'admin',
			preferredCurrency: 'USD',
			createdAt: NOW,
			updatedAt: NOW
		},
		session: {
			id: 'session-admin',
			userId: 'admin-1',
			expiresAt: FUTURE,
			ipAddress: null,
			userAgent: null,
			createdAt: NOW
		}
	}
};

function envelope(data, message = 'ok') {
	return { success: true, message, details: '', data, timestamp: NOW };
}

function errorEnvelope(message) {
	return { success: false, message, details: '', timestamp: NOW };
}

function send(res, status, body, headers = {}) {
	res.writeHead(status, { 'content-type': 'application/json', ...headers });
	res.end(JSON.stringify(body));
}

function readBody(req) {
	return new Promise((resolve) => {
		const chunks = [];
		req.on('data', (c) => chunks.push(c));
		req.on('end', () => resolve(Buffer.concat(chunks)));
	});
}

function accountByToken(req) {
	const auth = req.headers.authorization ?? '';
	const token = auth.replace(/^Bearer\s+/i, '');
	return Object.values(ACCOUNTS).find((a) => a.accessToken === token) ?? null;
}

function accountByRefreshCookie(req) {
	const cookie = req.headers.cookie ?? '';
	const match = cookie.match(/refresh_token=([^;\s]+)/);
	if (!match) return null;
	return Object.values(ACCOUNTS).find((a) => a.refreshToken === match[1]) ?? null;
}

function refreshSetCookie(account) {
	return `refresh_token=${account.refreshToken}; Path=/; HttpOnly; SameSite=Strict; Max-Age=2592000`;
}

/**
 * Tokens MCP creados durante la sesión del stub.
 *
 * Es el único estado mutable que guarda: el flujo que prueba la suite —crear un
 * token y verlo en la lista— no se puede montar con una fixture fija, porque lo
 * que hay que comprobar es justo que aparezca lo que se acaba de crear. El
 * proceso se levanta por corrida de Playwright, así que se vacía solo.
 */
const mcpTokens = [];

const server = createServer(async (req, res) => {
	const url = new URL(req.url, `http://127.0.0.1:${PORT}`);
	if (!url.pathname.startsWith(API_PREFIX)) {
		return send(res, 404, errorEnvelope('not found'));
	}
	const path = url.pathname.slice(API_PREFIX.length) || '/';
	const route = `${req.method} ${path}`;

	// ---- Public auth routes ----
	if (route === 'POST /auth/login') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		const account = ACCOUNTS[body.email];
		if (!account || body.password !== PASSWORD) {
			return send(res, 401, errorEnvelope('Credenciales incorrectas'));
		}
		return send(res, 200, envelope({ accessToken: account.accessToken }), {
			'set-cookie': refreshSetCookie(account)
		});
	}

	if (route === 'POST /auth/refresh') {
		const account = accountByRefreshCookie(req);
		if (!account) return send(res, 401, errorEnvelope('invalid refresh token'));
		return send(res, 200, envelope({ accessToken: account.accessToken }), {
			'set-cookie': refreshSetCookie(account)
		});
	}

	// ---- Everything below requires a valid access token ----
	const account = accountByToken(req);
	if (!account) {
		await readBody(req);
		return send(res, 401, errorEnvelope('invalid or missing token'));
	}

	if (route === 'GET /auth/session') {
		return send(res, 200, envelope({ user: account.user, session: account.session }));
	}
	if (route === 'POST /auth/logout') {
		return send(res, 200, envelope(null, 'logged out'));
	}
	if (route === 'GET /auth/sessions') {
		return send(
			res,
			200,
			envelope([
				{
					id: account.session.id,
					ipAddress: '127.0.0.1',
					userAgent: 'Playwright e2e',
					location: null,
					createdAt: NOW,
					lastActiveAt: NOW,
					expiresAt: FUTURE,
					current: true
				}
			])
		);
	}
	if (route === 'GET /auth/2fa') {
		return send(res, 200, envelope({ enabled: false, pendingSetup: false, recoveryCodesLeft: 0 }));
	}

	// ---- Tokens MCP ----
	//
	// El stub los guarda en memoria porque el flujo que interesa probar es de
	// dos pasos: crear un token y verlo aparecer en la lista. El secreto se
	// devuelve solo aquí, igual que en el backend real.
	if (route === 'GET /auth/mcp-tokens') {
		return send(res, 200, envelope(mcpTokens));
	}
	if (route === 'POST /auth/mcp-tokens') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');

		if (mcpTokens.some((t) => t.name.toLowerCase() === String(body.name ?? '').toLowerCase())) {
			return send(res, 409, errorEnvelope('you already have a token with that name'));
		}

		const created = {
			id: `mcp-${mcpTokens.length + 1}`,
			name: body.name,
			last4: 'a3f9',
			expiresAt: body.expiresInDays === 0 ? null : FUTURE,
			lastUsedAt: null,
			rotatedAt: null,
			createdAt: NOW,
			expired: false
		};
		mcpTokens.push(created);

		return send(
			res,
			201,
			envelope({ ...created, token: 'fnx_mcp_e2e-secreto' }, 'MCP token created')
		);
	}
	if (path.startsWith('/auth/mcp-tokens/') && path.endsWith('/rotate') && req.method === 'POST') {
		await readBody(req);
		const id = path.slice('/auth/mcp-tokens/'.length, -'/rotate'.length);
		const token = mcpTokens.find((t) => t.id === id);
		if (!token) return send(res, 404, errorEnvelope('mcp token not found'));

		token.rotatedAt = NOW;
		token.lastUsedAt = null;

		return send(res, 200, envelope({ ...token, token: 'fnx_mcp_e2e-rotado' }, 'MCP token rotated'));
	}
	if (path.startsWith('/auth/mcp-tokens/') && req.method === 'DELETE') {
		const id = path.slice('/auth/mcp-tokens/'.length);
		const index = mcpTokens.findIndex((t) => t.id === id);
		if (index === -1) return send(res, 404, errorEnvelope('mcp token not found'));

		mcpTokens.splice(index, 1);

		return send(res, 200, envelope(null, 'MCP token deleted'));
	}

	// ---- Users ----
	if (route === 'GET /users/me/preferences') {
		return send(
			res,
			200,
			envelope({ userId: account.session.userId, emailAlerts: true, weeklySummary: false })
		);
	}
	if (route === 'PATCH /users/me/preferences') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		return send(res, 200, envelope({ userId: account.session.userId, ...body }));
	}
	if (route === 'PATCH /users/me') {
		const body = JSON.parse((await readBody(req)).toString() || '{}');
		return send(res, 200, envelope({ ...account.user, ...body }));
	}
	if (path === '/users' && req.method === 'GET') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: Object.values(ACCOUNTS).map((a, i) => ({
					id: a.session.userId,
					name: a.user.name,
					email: a.user.email,
					emailVerified: a.user.emailVerified,
					createdAt: NOW,
					bannedAt: null,
					role: { name: a.user.role },
					index: i
				})),
				metaData: {
					currentPage: 1,
					usersForPage: 20,
					offset: 0,
					totalUsers: Object.keys(ACCOUNTS).length,
					totalPages: 1,
					previous: false,
					next: false
				}
			})
		);
	}
	if (route === 'GET /users/invitations') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: [
					{
						id: 'invite-1',
						email: 'invitada@finexia.test',
						name: 'Invitada',
						role: 'customer',
						status: 'pending',
						expiresAt: FUTURE,
						createdAt: NOW
					}
				]
			})
		);
	}
	if (route === 'GET /users/waitlist') {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		return send(
			res,
			200,
			envelope({
				items: [
					{
						id: 'wait-1',
						email: 'espera@finexia.test',
						status: 'pending',
						invitedAt: null,
						createdAt: NOW
					}
				]
			})
		);
	}
	// El backend responde 204 sin cuerpo; el mock no guarda estado, así que la
	// fila sigue ahí al recargar y lo que se comprueba es que la action llega.
	if (req.method === 'DELETE' && /^\/users\/waitlist\/[^/]+$/.test(path)) {
		if (account.user.role !== 'admin') return send(res, 403, errorEnvelope('forbidden'));
		res.writeHead(204);
		res.end();
		return;
	}

	// ---- Portfolios ----
	if (route === 'GET /portfolios/risks') {
		return send(res, 200, envelope(risks));
	}
	if (route === 'GET /portfolios/summary') {
		return send(res, 200, envelope(portfolioSummary(url.searchParams.get('currency') ?? 'USD')));
	}
	if (route === 'GET /portfolios/transactions') {
		return send(res, 200, envelope(transactions));
	}
	if (route === 'GET /portfolios/allocation') {
		return send(res, 200, envelope(allocation));
	}
	if (route === 'GET /portfolios/holdings') {
		return send(res, 200, envelope(assetHoldings));
	}
	if (route === 'GET /portfolios/growth') {
		return send(res, 200, envelope(growth));
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}\/growth$/.test(path)) {
		return send(res, 200, envelope(growthFor(path.split('/')[2])));
	}
	if (route === 'GET /portfolios/sources') {
		return send(res, 200, envelope(sources));
	}
	// El backend se niega a borrar una plataforma que todavía apuntan
	// posiciones: no las arrastra consigo, porque son el historial del dueño.
	// El stub reproduce las dos respuestas —409 con el motivo, 200 cuando no
	// queda nada que la apunte— para que el flujo de la UI se pruebe contra la
	// forma real y no contra un borrado que siempre funciona.
	if (req.method === 'DELETE' && /^\/portfolios\/sources\/[0-9a-f-]{36}$/.test(path)) {
		const source = sources.find((s) => s.id === path.split('/')[3]);
		if (!source) {
			return send(res, 404, errorEnvelope('Platform not found'));
		}
		if (source.investments > 0) {
			return send(res, 409, {
				success: false,
				message: 'Error deleting platform',
				details: `platform still has positions: ${source.investments} position(s), closed ones included, still reference it`,
				action: 'Could not delete platform',
				timestamp: NOW
			});
		}
		return send(res, 200, envelope(null, 'Platform deleted'));
	}
	if (route === 'GET /portfolios/assets') {
		const search = (url.searchParams.get('search') ?? '').toLowerCase();
		const filtered = search
			? assets.filter(
					(a) => a.ticker.toLowerCase().includes(search) || a.name.toLowerCase().includes(search)
				)
			: assets;
		return send(res, 200, envelope(filtered));
	}
	// ---- Datos de mercado (BYO-key) ----
	if (route === 'GET /market/credentials') {
		return send(res, 200, envelope(marketCredentials));
	}
	// Repide el precio de un activo al proveedor que nombra el cuerpo. El stub
	// no guarda nada —como el resto de sus escrituras— así que una recarga
	// vuelve a traer el precio del fixture: lo que se prueba aquí es que la UI
	// nombra al proveedor elegido y enseña su respuesta, no la persistencia.
	if (req.method === 'POST' && /^\/market\/assets\/[0-9a-f-]{36}\/refresh$/.test(path)) {
		const assetId = path.split('/')[3];
		const holding = assetHoldings.find((h) => h.assetId === assetId);
		if (!holding) {
			return send(res, 404, errorEnvelope('asset not found'));
		}

		const body = JSON.parse((await readBody(req)).toString() || '{}');
		const provider = body.provider;

		// Alpha Vantage no cubre el fixture de cripto: es el camino de error que
		// distingue «este proveedor no lo tiene» de «la clave no sirve», y la UI
		// tiene que enseñarlo tal cual viene.
		if (provider === 'alphavantage' && holding.assetType === 'crypto') {
			return send(res, 400, {
				success: false,
				message: 'Could not refresh the price',
				details: 'Este proveedor no tiene datos de este activo. Prueba con otra de tus claves.',
				action: 'Could not refresh the price',
				timestamp: NOW
			});
		}

		return send(
			res,
			200,
			envelope(
				{
					assetId,
					ticker: holding.ticker,
					price: holding.marketPrice,
					source: provider,
					fetchedAt: NOW
				},
				'Price updated'
			)
		);
	}
	if (route === 'GET /exchange-rates') {
		return send(res, 200, envelope(exchangeRates));
	}
	// Sin paginar y abierto a cualquier usuario, a diferencia del listado de
	// arriba: lo pide el dashboard para enseñar con qué se convierten sus cifras.
	if (route === 'GET /exchange-rates/latest') {
		return send(res, 200, envelope(exchangeRates));
	}
	// Devuelve solo lo que publica el feed, que es lo que el backend responde:
	// las tasas que acaba de reescribir, no la tabla entera.
	if (route === 'POST /exchange-rates/refresh') {
		return send(
			res,
			200,
			envelope(
				exchangeRates.filter((r) => r.source === 'dolarapi'),
				'exchange rates refreshed'
			)
		);
	}
	if (route === 'POST /portfolios/entries') {
		await readBody(req);
		return send(res, 201, envelope({ id: IDS.entry }, 'entry created'));
	}
	if (route === 'POST /portfolios/transactions/import/preview') {
		await readBody(req);
		return send(res, 200, envelope(importPreview));
	}
	if (
		req.method === 'GET' &&
		/^\/portfolios\/[0-9a-f-]{36}\/assets\/[^/]+\/transactions$/.test(path)
	) {
		const symbol = decodeURIComponent(path.split('/')[4]);
		const rows = transactions.filter((t) => t.assetTicker === symbol);
		return send(
			res,
			200,
			envelope({ data: rows, total: rows.length, page: 1, limit: 20, totalPages: 1 })
		);
	}
	// El backend responde 200 con `data: null`; la posición la recalcula la
	// base, así que no hay nada que devolver.
	if (req.method === 'DELETE' && /^\/portfolios\/transactions\/[0-9a-f-]{36}$/.test(path)) {
		return send(res, 200, envelope(null, 'transaction deleted'));
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}\/top-transaction$/.test(path)) {
		const top = topTransaction(path.split('/')[2]);
		if (!top) return send(res, 404, errorEnvelope('no transactions'));
		return send(res, 200, envelope(top));
	}
	if (req.method === 'GET' && /^\/portfolios\/[0-9a-f-]{36}$/.test(path)) {
		const portfolio = PORTFOLIOS.find((p) => p.id === path.split('/')[2]);
		if (!portfolio) return send(res, 404, errorEnvelope('portfolio not found'));
		return send(
			res,
			200,
			envelope({
				id: portfolio.id,
				userId: account.session.userId,
				name: portfolio.name,
				description: portfolio.description,
				type: portfolio.type,
				baseCurrency: 'USD',
				isDefault: portfolio.isDefault,
				riskId: portfolio.riskId,
				riskName: portfolio.riskName,
				createdAt: NOW,
				updatedAt: NOW,
				holdings: portfolio.holdings
			})
		);
	}

	await readBody(req);
	return send(res, 404, errorEnvelope(`no mock for ${route}`));
});

// Solo escucha cuando se ejecuta como programa (así lo arranca Playwright);
// importarlo desde un test trae las fixtures sin abrir un puerto.
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
	server.listen(PORT, () => {
		console.log(`mock backend listening on http://127.0.0.1:${PORT}${API_PREFIX}`);
	});
}
