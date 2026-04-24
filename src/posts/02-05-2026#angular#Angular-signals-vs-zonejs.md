<p align="middle">
  <img src="https://angular.dev/assets/icons/logo.svg" align="middle" title="Angular" alt="Angular logo" width="120px">
</p>

Angular Signals are stable now, signal-based inputs and outputs are here, and zoneless mode finally feels real.

But one thing is worth saying first: Angular did not become reactive only now.

## Angular was already reactive

A lot of people coming from newer frameworks look at Signals and think "now Angular is reactive". Not really.

Angular has always been reactive in its own way. For years, that reactivity was powered mostly by `zone.js`.

The simple explanation is this: `zone.js` watches async things happening in the browser, like clicks, timers, HTTP callbacks and promises. When something happens, Angular runs change detection again and updates the screen.

That worked. It still works. The problem is that it is broad. Angular often needs to check much more than what actually changed.

So yes, Angular was already reactive. It was just a more global and less precise kind of reactivity.

## What Signals improve

Signals make that reactivity explicit.

Instead of Angular asking "did anything change somewhere?", it can now track exactly what a component depends on and update only the parts that need to move.

```typescript
import { Component, computed, input, output, signal } from '@angular/core';

@Component({
  selector: 'app-product-card',
  standalone: true,
  template: `
    <h2>{{ title() }}</h2>
    <button (click)="addToCart.emit(product())">
      Add to cart
    </button>
  `
})
export class ProductCardComponent {
  product = input.required<Product>();
  addToCart = output<Product>();
  quantity = signal(1);
  title = computed(() => `${this.product().name} x${this.quantity()}`);
}
```

This is the part I like most. The component becomes easier to read. Local state feels cleaner. Derived state is obvious. And APIs like `input()`, `output()` and `model()` make the component model feel much more cohesive than before.

## The real win is zoneless

To me, this is the bigger story.

Signals are not just a nicer state API. They are part of what allows Angular to stop depending on `zone.js` for everything.

When the framework knows exactly what changed, it does not need to monkey-patch browser APIs and trigger wide change detection cycles all the time. That is where the performance win starts to become really interesting.

And combined with the newer rendering and hydration direction in Angular, the framework feels much more modern than it did a few versions ago.

The final advantage is simple: you can build an Angular app that keeps the framework strengths, but now with a much clearer path to a `zoneless` app.

## RxJS still matters

Going all-in on Signals does not mean abandoning RxJS.

I still use RxJS for HTTP flows, WebSockets, debouncing, cancellation and stream composition. Signals are amazing for local state and UI reactivity. RxJS is still better for async orchestration.

Best of both worlds.

## Final thoughts

Signals did not make Angular reactive. Angular was already reactive, just with `zone.js` doing a lot of heavy lifting behind the scenes.

What Signals bring is a more explicit, more predictable and more performant model. And for me, the best part of that story is not the syntax. It is where this leads: Angular with a real path to `zoneless`.
