<script lang="ts">
	const stats = [
		{ value: '47,000+', label: 'inversiones rastreadas' },
		{ value: '15', label: 'brokers compatibles' },
		{ value: '∞', label: 'sin límite de activos' }
	];

	const testimonials = [
		{
			quote: 'Finalmente veo todo junto. Pasé de estar ansioso a tener control real sobre mi dinero.',
			author: 'Andrés',
			age: '34',
			context: 'Inversionista con 4 brokers'
		},
		{
			quote: 'No es un juguete financiero. Es una herramienta profesional que se siente accesible.',
			author: 'María',
			age: '28',
			context: 'Principiante en inversiones'
		}
	];
</script>

<section id="testimonios" class="mx-auto w-full max-w-7xl px-6 py-20">
	<div class="mb-16 space-y-4 text-center">
		<p class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">Confían en FINEXIA</p>
		<h2 class="font-display text-4xl font-bold leading-tight text-[#ffffff]">
			Inversores como tú ya tienen control
		</h2>
	</div>

	<div class="mb-16 grid gap-8 md:grid-cols-3">
		{#each stats as stat (stat.label)}
			<div class="stat-card">
				<p class="font-display text-4xl font-bold text-[#d4af37]">{stat.value}</p>
				<p class="mt-2 text-sm text-[#a0a0a0]">{stat.label}</p>
			</div>
		{/each}
	</div>

	<div class="grid gap-6 md:grid-cols-2">
		{#each testimonials as testimonial (testimonial.author)}
			<div class="testimonial-card">
				<div class="mb-4 text-3xl text-[#d4af37]">❝</div>
				<p class="mb-6 text-lg text-[#e0e0e0]">{testimonial.quote}</p>
				<div class="border-t border-[rgba(212,175,55,0.1)] pt-4">
					<p class="font-bold text-[#d4af37]">{testimonial.author}, {testimonial.age}</p>
					<p class="text-xs text-[#a0a0a0]">{testimonial.context}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<style>
	.stat-card {
		border: 1px solid rgba(212, 175, 55, 0.2);
		background: linear-gradient(135deg, #1a1f2e 0%, #141a24 100%);
		padding: 2rem;
		border-radius: 16px;
		text-align: center;
		backdrop-filter: blur(10px);
		transition: all 0.3s ease;
	}

	.stat-card:hover {
		border-color: rgba(212, 175, 55, 0.4);
		transform: translateY(-4px);
		box-shadow: 0 12px 40px rgba(212, 175, 55, 0.15);
	}

	.testimonial-card {
		border: 1px solid rgba(212, 175, 55, 0.2);
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.6) 0%, rgba(20, 26, 36, 0.6) 100%);
		padding: 2rem;
		border-radius: 16px;
		backdrop-filter: blur(10px);
		transition: all 0.3s ease;
	}

	.testimonial-card:hover {
		border-color: rgba(212, 175, 55, 0.4);
		transform: translateY(-4px);
	}

	@media (max-width: 768px) {
		.stat-card,
		.testimonial-card {
			padding: 1.5rem;
		}
	}
</style>
