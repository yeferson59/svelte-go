import { expect, test } from '@playwright/test';
import { login } from './helpers';

/*
 * La vista consolidada de activos: una fila por activo con lo que el usuario
 * tiene de él sumando todos sus portafolios, y la barra que dice cómo está
 * repartido el dinero entre ellos.
 */
test.describe('mis activos', () => {
	test('lists every asset with its type, units and weight', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await expect(page.getByRole('heading', { level: 1 })).toContainText('Mis activos');

		const table = page.getByRole('table');
		await expect(table).toBeVisible();

		// Una fila por activo del fixture, no una por posición: BTC vive en el
		// portafolio de cripto y aparece una sola vez, con sus unidades.
		const btc = table.getByRole('row').filter({ hasText: 'Bitcoin' });
		await expect(btc).toHaveCount(1);
		await expect(btc).toContainText('Cripto');
		await expect(btc).toContainText('0,15 uds');

		// Las diez posiciones del fixture son diez activos distintos.
		await expect(table.locator('tbody tr')).toHaveCount(10);
	});

	// Los anchos de la barra son proporcionales al valor, así que su punto medio
	// es la mitad del dinero: lo que se lee es cuántos activos caben a su
	// izquierda, y eso es lo que dice el pie.
	test('the concentration bar says how much of the money sits in the top assets', async ({
		page
	}) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await expect(page.getByRole('heading', { name: 'Cómo está repartido' })).toBeVisible();

		const caption = page.locator('section[aria-labelledby="concentration-title"] p');
		await expect(caption).toContainText('La marca señala la mitad de tu valor');

		// Señalar una fila enciende su franja y el pie pasa a describirla: es lo
		// que hace que la barra y la lista sean el mismo objeto.
		await page.getByRole('row').filter({ hasText: 'Bitcoin' }).hover();
		await expect(caption).toContainText('BTC');
		await expect(caption).toContainText('11,3%');
	});

	// La lista se pagina, así que sin buscador encontrar un activo era saber en
	// qué hoja cayó.
	test('filters the list down to the asset you are looking for', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await page.getByPlaceholder('Buscar un activo').fill('bitcoin');
		await expect(page.locator('tbody tr')).toHaveCount(1);
		await expect(page.locator('tbody tr')).toContainText('BTC');

		await page.getByPlaceholder('Buscar un activo').fill('no existe');
		await expect(page.getByText('Ningún activo se llama así.')).toBeVisible();

		await page.getByRole('button', { name: 'Ver todos' }).click();
		await expect(page.locator('tbody tr')).toHaveCount(10);
	});

	/*
	 * Pulsar un activo abre lo que la lista no cabe diciendo: de qué clave salió
	 * su precio, y con cuál volver a pedirlo. El botón «Sincronizar» de Ajustes
	 * recorre todas las tenencias y elige el proveedor por su cuenta; aquí se
	 * elige a mano y se pregunta por un solo activo.
	 */
	test('opens an asset to see where its price came from and re-ask a provider', async ({
		page
	}) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await page.getByRole('button', { name: /precio de BTC/ }).click();

		const panel = page.getByRole('dialog');
		await expect(panel).toBeVisible();
		await expect(panel.getByText(/Lo trajo Finnhub hace/)).toBeVisible();

		// Viene marcado el proveedor que trajo el precio que se está mirando:
		// repreguntar a quien contestó es lo que se quiere casi siempre.
		await expect(panel.getByRole('radio', { name: /Finnhub/ })).toBeChecked();

		await panel.getByRole('button', { name: 'Actualizar precio' }).click();
		await expect(panel.getByText(/Finnhub respondió/)).toBeVisible();
	});

	// Que un proveedor no cubra un activo no es que la clave falle, y son dos
	// arreglos distintos: el backend lo dice con esas palabras y la UI lo pone
	// tal cual, en el activo sobre el que se preguntó.
	test('says when the chosen provider has no data for the asset', async ({ page }) => {
		await login(page);
		await page.goto('/dashboard/assets');

		await page.getByRole('button', { name: /precio de BTC/ }).click();

		const panel = page.getByRole('dialog');
		await panel.getByRole('radio', { name: /Alpha Vantage/ }).check();
		await panel.getByRole('button', { name: 'Actualizar precio' }).click();

		await expect(panel.getByText(/no tiene datos de este activo/)).toBeVisible();
	});
});
