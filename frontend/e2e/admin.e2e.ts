import { expect, test } from '@playwright/test';
import { ADMIN_EMAIL, login } from './helpers';

test.describe('admin', () => {
	test('lists registered users for an admin', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/users');

		await expect(page.getByRole('heading', { name: 'Usuarios registrados' })).toBeVisible();
		await expect(page.getByText('user@finexia.test').first()).toBeVisible();
		await expect(page.getByRole('heading', { name: 'Invitaciones pendientes' })).toBeVisible();
		await expect(page.getByText('espera@finexia.test').first()).toBeVisible();
	});

	// La baja de la lista de espera es la única acción destructiva de esa tabla;
	// el mock no guarda estado, así que lo que se comprueba es que la action
	// llega al endpoint y vuelve sin error por fila.
	test('removes an entry from the waitlist', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/users');

		const row = page.locator('tr', { hasText: 'espera@finexia.test' });
		await row.getByRole('button', { name: 'Eliminar' }).click();

		await expect(page.getByRole('heading', { name: 'Lista de espera' })).toBeVisible();
		await expect(row.locator('.row-error')).toHaveCount(0);
	});

	test('redirects non-admin users back to the dashboard', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/admin/users');
		await expect(page).toHaveURL(/\/dashboard$/);
	});

	test('lists the shared asset catalogue with its manual price', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/assets');

		const row = page.locator('tr', { hasText: 'AAPL' });
		await expect(row.getByText('Apple Inc.')).toBeVisible();
		await expect(row.getByText('$214.35')).toBeVisible();
		// El precio manual se edita en la propia fila.
		await expect(row.locator('input[name="price"]')).toHaveValue('214.35');
	});

	test('opens the asset create and import forms', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/assets');

		await page.getByRole('button', { name: 'Nuevo activo' }).click();
		await expect(page.getByRole('heading', { name: 'Nuevo activo' })).toBeVisible();
		await expect(page.locator('#ticker')).toBeVisible();

		// Cada alta es un modal, así que hay que cerrar uno para abrir el otro.
		// Escape lo trae el `<dialog>` nativo.
		await page.keyboard.press('Escape');
		await expect(page.getByRole('heading', { name: 'Nuevo activo' })).not.toBeVisible();

		await page.getByRole('button', { name: 'Importar CSV/Excel' }).click();
		await expect(
			page.getByRole('heading', { name: 'Importar activos desde CSV/Excel' })
		).toBeVisible();
		await expect(page.locator('input[type="file"][name="file"]')).toBeVisible();
	});

	// La ficha completa: lo que la columna de precio no alcanza —el ticker, el
	// nombre, el tipo, la moneda y quién ve la fila— y que antes obligaba a
	// volver a darla de alta con el mismo ticker para corregir una errata.
	test('edits an asset from the catalogue', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/assets');

		const row = page.locator('tr', { hasText: 'AAPL' });
		await row.getByRole('button', { name: 'Editar' }).click();

		await expect(page.getByRole('heading', { name: 'Editar activo' })).toBeVisible();
		// La ficha llega llena: es una corrección, no un alta.
		await expect(page.locator('#edit-ticker')).toHaveValue('AAPL');
		await expect(page.locator('#edit-name')).toHaveValue('Apple Inc.');
		await expect(page.locator('#edit-type')).toHaveValue('stock');
		await expect(page.locator('#edit-currency')).toHaveValue('USD');
		// El precio nuevo se pide en blanco: dejarlo así conserva el guardado.
		await expect(page.locator('#edit-price')).toHaveValue('');

		await page.locator('#edit-name').fill('Apple Inc. (corregido)');
		await page.getByRole('button', { name: 'Guardar cambios' }).click();

		await expect(page.getByRole('heading', { name: 'Editar activo' })).not.toBeVisible();
		await expect(row.getByText('Cambios guardados')).toBeVisible();
	});

	test('lists the exchange rates and opens the create form', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/exchange-rates');

		const row = page.locator('tr', { hasText: 'USD/COP' });
		await expect(row.getByText('4,123.456789')).toBeVisible();
		await expect(row.locator('input[name="rate"]')).toHaveValue('4123.456789');

		// El origen distingue las filas que el feed reescribe cada hora de las
		// que sobreviven a un refresco, que es lo que un admin necesita saber
		// antes de corregir una a mano.
		await expect(row.getByText('TRM (automática)')).toBeVisible();
		await expect(page.locator('tr', { hasText: 'EUR/USD' }).getByText('Manual')).toBeVisible();

		await page.getByRole('button', { name: 'Nueva tasa' }).click();
		await expect(page.getByRole('heading', { name: 'Nueva tasa de cambio' })).toBeVisible();
		await expect(page.locator('#fromCurrency')).toBeVisible();
	});

	// El resalte del hover se pinta sobre el fondo de cada `td`, así que una
	// celda más baja que su fila lo deja mordido. Es lo que hacía un
	// `display: flex` sobre la primera celda: en flex un `td` deja de ser celda
	// de tabla y se encoge a su contenido. Se comprueba la geometría porque el
	// síntoma es visual pero la causa se mide.
	test('every cell fills its row, so the hover highlight is unbroken', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/exchange-rates');

		const mismatches = await page.evaluate(() =>
			[...document.querySelectorAll('tbody tr')].flatMap((row) => {
				const rowHeight = row.getBoundingClientRect().height;

				return [...row.querySelectorAll('td')]
					.filter((td) => Math.abs(td.getBoundingClientRect().height - rowHeight) > 0.5)
					.map((td) => `${td.className}: ${td.getBoundingClientRect().height} vs ${rowHeight}`);
			})
		);

		expect(mismatches).toEqual([]);
	});

	test('refreshes the shared rates from the public feed', async ({ page }) => {
		await login(page, ADMIN_EMAIL);
		await page.goto('/dashboard/admin/exchange-rates');

		await page.getByRole('button', { name: 'Actualizar desde el feed' }).click();

		await expect(page.getByText('1 tasa actualizada desde el feed.')).toBeVisible();
	});
});
