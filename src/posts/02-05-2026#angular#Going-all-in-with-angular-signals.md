<p align="middle">
  <img src="https://angular.dev/assets/icons/logo.svg" align="middle" title="Angular" alt="Angular logo" width="120px">
</p>

Angular Signals is finally stable. We now have signal-based inputs and outputs, zoneless mode is stable, and the whole component model feels much more cohesive than it did in the early releases.

## The full signal-based component

The component model changed significantly. Here's what a modern Angular component looks like today:

```typescript
import { Component, computed, input, output } from '@angular/core';

@Component({
  selector: 'app-product-card',
  standalone: true,
  template: `
    <h2>{{ product().name }}</h2>
    <button (click)="addToCart.emit(product())">Add to cart</button>
  `
})
export class ProductCardComponent {
  product = input.required<Product>();
  addToCart = output<Product>();
  label = computed(() => this.product().name.toUpperCase());
}
```

`input()` and `output()` are signal-based now. `input.required()` gives you a compile-time guarantee that the parent has to pass that value. No more `@Input() product!: Product` with the non-null assertion hack.

## State management with signals

For component state, signals replaced almost all my uses of local variables combined with `markForCheck()` or `detectChanges()`. Before I'd do things like:

```typescript
export class OldComponent implements OnInit {
  loading = false;
  items: Item[] = [];

  ngOnInit() {
    this.loading = true;
    this.service.getItems().subscribe({
      next: items => {
        this.items = items;
        this.loading = false;
      }
    });
  }
}
```

Now with signals:

```typescript
import { Component, signal, computed, inject } from '@angular/core';

export class NewComponent {
  private service = inject(ItemService);

  items = signal<Item[]>([]);
  loading = signal(false);

  hasItems = computed(() => this.items().length > 0);

  loadItems() {
    this.loading.set(true);
    this.service.getItems().subscribe({
      next: items => {
        this.items.set(items);
        this.loading.set(false);
      }
    });
  }
}
```

The template can now react to `loading()`, `items()` and `hasItems()` precisely. Angular only updates the DOM nodes that actually depend on a changed signal.

## model() — two-way binding the right way

This one was a game changer for form-like components. `model()` is a signal that's both readable and writable from outside:

```typescript
import { Component, model } from '@angular/core';

@Component({
  selector: 'app-toggle',
  standalone: true,
  template: `
    <button (click)="toggle()" [class.active]="checked()">
      {{ checked() ? 'ON' : 'OFF' }}
    </button>
  `
})
export class ToggleComponent {
  checked = model(false);

  toggle() { this.checked.update(v => !v); }
}
```

Using it from the parent:

```html
<app-toggle [(checked)]="isEnabled" />
```

Two-way binding, but clean. No need for `EventEmitter` + `@Output` + manually naming things with the `Change` suffix.

## Zoneless Angular

The big thing in Angular 19 is that zoneless mode is now stable and I've been running apps without Zone.js. The bundle size reduction is real.

To enable it:

```typescript
// main.ts
bootstrapApplication(AppComponent, {
  providers: [provideExperimentalZonelessChangeDetection()]
});
```

And remove `zone.js` from your `polyfills` in `angular.json`. 

With signals driving change detection, Angular knows exactly what changed and when. No more monkey-patching browser APIs. The performance difference in long-running apps is noticeable, especially with a lot of components on screen at the same time.

## What I still use RxJS for

Even going all-in on signals, I haven't abandoned RxJS. HTTP calls, WebSocket streams, complex event debouncing — RxJS still handles those much better. The `toSignal()` bridge makes it seamless:

```typescript
export class SearchComponent {
  private http = inject(HttpClient);
  query = signal('');
  results = toSignal(
    toObservable(this.query).pipe(
      switchMap(q => q ? this.http.get<Result[]>(`/api/search?q=${q}`) : of([]))
    ),
    { initialValue: [] }
  );
}
```

Best of both worlds. Signals for state, RxJS for async orchestration.

## Final thoughts

Signals solved the biggest pain points I had with Angular. Change detection is now predictable. Components are easier to reason about. And with zoneless mode stable, the performance story is genuinely competitive with any other framework out there.

If you haven't migrated to signal-based inputs and outputs yet, the Angular CLI migration schematics make it pretty straightforward. Highly recommend.

See you on the next post 🤙
