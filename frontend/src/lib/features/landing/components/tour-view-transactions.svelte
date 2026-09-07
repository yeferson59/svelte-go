<script lang="ts">
	/*
	 * Maqueta de `/dashboard/transactions`: el libro de movimientos.
	 *
	 * Enseñaba dos pantallas a la vez —el asistente de importación, que vive en
	 * otra ruta, y un historial de cápsulas de color por tipo— y el asistente
	 * tenía cuatro pasos donde el real tiene tres. Hoy es la tabla que hay: qué
	 * pasó, sobre qué activo, cuándo, y la cantidad y el precio con los que se
	 * comprueba el total. El color de la fila se lo lleva el signo del importe,
	 * no una cápsula por tipo.
	 */
	import { TOUR_TRANSACTIONS } from '../product-tour';
</script>

<div class="transactions">
	<div class="page-head">
		<div>
			<p class="mk-page-title">Transacciones</p>
			<p class="mk-page-sub">
				Todo lo que has registrado, del movimiento más reciente al más antiguo.
			</p>
		</div>
		<span class="mk-btn">Importar desde Excel</span>
	</div>

	<table class="mk-table">
		<thead>
			<tr>
				<th class="col-kind">Movimiento</th>
				<th class="col-asset">Activo</th>
				<th class="col-date">Fecha</th>
				<th class="mk-num col-qty">Cantidad</th>
				<th class="mk-num col-price">Precio</th>
				<th class="mk-num">Total</th>
			</tr>
		</thead>
		<tbody>
			{#each TOUR_TRANSACTIONS as tx (tx.ticker + tx.date)}
				<tr>
					<th class="kind">
						{tx.kind}
						<!-- La nota es lo único que escribió el usuario, así que se ve. -->
						{#if tx.note}<span class="mk-detail">{tx.note}</span>{/if}
					</th>
					<td class="col-asset">
						{tx.asset}
						<span class="ticker">{tx.ticker}</span>
					</td>
					<td class="col-date date">{tx.date}</td>
					<td class="mk-num col-qty">{tx.qty}</td>
					<td class="mk-num col-price muted">{tx.price}</td>
					<td class="mk-num total">{tx.total}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 20px;
		padding-bottom: 16px;
		margin-bottom: 18px;
		border-bottom: 1px solid var(--border);
	}

	.kind {
		font-weight: 500;
		white-space: nowrap;
	}

	.col-asset {
		width: 32%;
	}

	/* El nombre manda y el ticker va debajo: en una línea, «Vanguard FTSE
	   All-World UCITS ETF (VWCE)» empuja las cifras fuera de la pantalla. */
	.ticker {
		display: block;
		margin-top: 2px;
		font-family: var(--font-mono);
		font-size: 8.5px;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.date {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.muted {
		color: var(--text-muted);
	}

	.total {
		font-weight: 600;
	}

	@media (max-width: 780px) {
		.col-price {
			display: none;
		}
	}

	@media (max-width: 620px) {
		.page-head {
			flex-direction: column;
			gap: 12px;
		}

		.col-date,
		.col-qty {
			display: none;
		}
	}
</style>
