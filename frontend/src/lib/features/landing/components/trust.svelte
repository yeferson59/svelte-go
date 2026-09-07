<script lang="ts">
	/*
	 * Garantías de seguridad. Cada una corresponde a algo que la aplicación hace
	 * hoy —2FA, cierre de sesiones, modo oculto, exportación en XLSX—, no a una
	 * promesa a futuro: la landing no debe prometer lo que el dashboard no tiene.
	 *
	 * La cabecera va a dos columnas (titular y argumento sobre un filete), la
	 * tercera forma distinta de abrir sección de la página.
	 */
	const guarantees = [
		{
			icon: 'lock',
			title: 'Nunca pedimos credenciales',
			body: 'Finexia no se conecta a tu broker, exchange ni banco. No hay contraseñas de terceros que guardar, así que no hay nada que se nos pueda escapar.'
		},
		{
			icon: 'shield',
			title: 'Verificación en dos pasos',
			body: 'Activa el segundo factor con tu app de autenticación y guarda los códigos de recuperación por si pierdes el teléfono.'
		},
		{
			icon: 'devices',
			title: 'Tus sesiones, bajo control',
			body: 'Consulta desde qué dispositivos hay sesión abierta y ciérralas una a una o todas de golpe cuando quieras.'
		},
		{
			icon: 'eye',
			/* El nombre que lleva el botón del panel. «Modo oculto» es como se llama
			   por dentro, y nadie lo lee en ninguna pantalla. */
			title: 'Ocultar los importes',
			body: 'Un clic y todos los importes quedan enmascarados. Puedes abrir tu panel en un café o en una reunión sin enseñar tus cifras.'
		},
		{
			icon: 'download',
			title: 'Tus datos salen contigo',
			body: 'Descarga tu resumen, tus movimientos y tu informe de riesgo en XLSX cuando te apetezca. Sin trámites ni permisos.'
		},
		{
			icon: 'pencil',
			title: 'Tú eres la única fuente',
			body: 'Todo lo que aparece en el panel lo has escrito tú. Nada se añade, se estima ni se comparte por detrás.'
		}
	];
</script>

<section class="wrap block" id="seguridad">
	<div class="sec-split reveal">
		<div>
			<div class="eyebrow">Seguridad y control</div>
			<h2 class="sec-title">Menos accesos,<br />menos superficie de riesgo</h2>
		</div>
		<p class="sec-desc">
			La forma más segura de tratar las llaves de tus cuentas es no pedirlas nunca. Sobre esa base
			está construido el resto.
		</p>
	</div>

	<div class="trust-grid">
		{#each guarantees as item (item.title)}
			<div class="titem reveal">
				<span class="ticon" aria-hidden="true">
					<svg
						width="17"
						height="17"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="1.7"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						{#if item.icon === 'lock'}
							<rect x="3" y="11" width="18" height="11" rx="2" />
							<path d="M7 11V7a5 5 0 0 1 10 0v4" />
						{:else if item.icon === 'shield'}
							<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
							<path d="m9 12 2 2 4-4" />
						{:else if item.icon === 'devices'}
							<rect x="2" y="4" width="14" height="10" rx="2" />
							<path d="M6 18h6" />
							<rect x="18" y="10" width="4" height="10" rx="1" />
						{:else if item.icon === 'eye'}
							<path d="M2 12s3.6-7 10-7 10 7 10 7-3.6 7-10 7-10-7-10-7z" />
							<circle cx="12" cy="12" r="3" />
							<path d="m3 3 18 18" />
						{:else if item.icon === 'download'}
							<path d="M12 3v12" />
							<path d="m7 11 5 5 5-5" />
							<path d="M4 19h16" />
						{:else}
							<path d="M12 20h9" />
							<path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" />
						{/if}
					</svg>
				</span>
				<h3>{item.title}</h3>
				<p>{item.body}</p>
			</div>
		{/each}
	</div>
</section>

<style>
	.trust-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 1px;
		background: var(--border);
		border: 1px solid var(--border);
		border-radius: 12px;
		overflow: hidden;
	}

	.titem {
		padding: 30px 26px;
		background: var(--bg);
		transition: background 0.2s;
	}

	.titem:hover {
		background: rgba(255, 255, 255, 0.022);
	}

	.ticon {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 9px;
		border: 1px solid rgba(212, 145, 42, 0.22);
		background: rgba(212, 145, 42, 0.08);
		color: var(--amber-light);
		margin-bottom: 18px;
	}

	.titem h3 {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 17px;
		letter-spacing: -0.01em;
		margin-bottom: 9px;
	}

	.titem p {
		font-size: 13.5px;
		color: var(--text-muted);
		line-height: 1.65;
		font-weight: 300;
		text-wrap: pretty;
	}

	@media (max-width: 900px) {
		.trust-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 600px) {
		.trust-grid {
			grid-template-columns: minmax(0, 1fr);
		}

		.titem {
			padding: 24px 20px;
		}
	}
</style>
