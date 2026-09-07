<script lang="ts">
	/*
	 * El extracto de la maqueta del resumen: los últimos movimientos, como en el
	 * panel. El signo del importe basta para saber si el dinero entró o salió,
	 * así que ninguna fila lleva icono ni cuadro de color.
	 *
	 * Interno de la feature.
	 */
	import { TOUR_ACTIVITY } from '../product-tour';
</script>

<div class="activity">
	<div class="head">
		<span class="mk-h2">Movimientos</span>
		<span class="link">Ver todos</span>
	</div>

	<ul class="rows">
		{#each TOUR_ACTIVITY as item (item.asset + item.date)}
			<li>
				<span class="what">
					<b>{item.kind}</b>
					<em>{item.asset}</em>
				</span>
				<span class="figures">
					<b class="mk-num" class:mk-up={item.incoming}>{item.amount}</b>
					<em>{item.date}</em>
				</span>
			</li>
		{/each}
	</ul>

	<span class="link statement">Descargar el extracto</span>
</div>

<style>
	.activity {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
	}

	.link {
		font-size: 9.5px;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.statement {
		align-self: flex-start;
		margin-top: 12px;
	}

	.rows {
		display: flex;
		flex-direction: column;
		gap: 3px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* La única superficie levantada de la página, y por eso significa algo: cada
	   fila es un objeto suelto del extracto, no una sección. */
	.rows li {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
		padding: 6px 8px;
		border-radius: 7px;
		background: rgba(255, 255, 255, 0.04);
	}

	.what {
		min-width: 0;
	}

	.what b,
	.figures b {
		display: block;
		font-size: 10px;
		font-weight: 400;
		color: var(--text);
	}

	.what em,
	.figures em {
		display: block;
		margin-top: 1px;
		font-style: normal;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.what em {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.figures {
		flex-shrink: 0;
		text-align: right;
	}

	.figures em {
		font-family: var(--font-mono);
	}
</style>
