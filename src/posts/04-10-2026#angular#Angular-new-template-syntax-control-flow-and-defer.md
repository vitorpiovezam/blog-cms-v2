<p align="middle">
  <img src="https://angular.dev/assets/icons/logo.svg" align="middle" title="Angular" alt="Angular logo" width="120px">
</p>

There's something that bugged me about Angular templates for years: `*ngIf`, `*ngFor`, `*ngSwitch`. They always felt a bit awkward. The asterisk syntax, having to import `CommonModule` just to use a conditional — small things that added up.

Angular 17 shipped a brand new built-in template syntax for control flow and lazy rendering. It's been stable for a while now and I've fully adopted it. Let me show you why it's so much better.

## The new @if block

Before:

```html
<div *ngIf="user; else loading">
  <p>Welcome, {{ user.name }}</p>
</div>
<ng-template #loading>
  <p>Loading...</p>
</ng-template>
```

After:

```html
@if (user) {
  <p>Welcome, {{ user.name }}</p>
} @else {
  <p>Loading...</p>
}
```

The `ng-template` workaround for the else block was always a bit of a hack. The new syntax is just... readable. Any developer coming from any language can understand this immediately.

You also get `@else if`:

```html
@if (status === 'loading') {
  <app-spinner />
} @else if (status === 'error') {
  <app-error [message]="errorMessage()" />
} @else {
  <app-content [data]="data()" />
}
```

## The new @for block

`*ngFor` had a known footgun: forgetting `trackBy` in large lists could kill your performance because Angular would re-render the whole list on any change.

The new `@for` makes `track` mandatory:

```html
@for (item of items(); track item.id) {
  <app-item-card [item]="item" />
} @empty {
  <p>No items found.</p>
}
```

Two things I love here:
1. `track` is required — you can't forget it
2. `@empty` block built right in, no more `*ngIf="items.length === 0"` alongside the loop

## The new @switch block

The old `[ngSwitch]` + `*ngSwitchCase` was genuinely ugly. Compare:

```html
<!-- Before -->
<div [ngSwitch]="userRole">
  <app-admin *ngSwitchCase="'admin'" />
  <app-editor *ngSwitchCase="'editor'" />
  <app-viewer *ngSwitchDefault />
</div>

<!-- After -->
@switch (userRole()) {
  @case ('admin') { <app-admin /> }
  @case ('editor') { <app-editor /> }
  @default { <app-viewer /> }
}
```

Much more readable. Also notice no `CommonModule` import needed — this syntax is built into the Angular compiler itself.

## @defer — lazy loading UI blocks

This one is the real headline feature. `@defer` lets you lazy-load parts of your template, including the components inside them, based on conditions you define.

```html
@defer (on viewport) {
  <app-comments [postId]="postId()" />
} @placeholder {
  <div class="comments-placeholder">Comments loading...</div>
} @loading (minimum 500ms) {
  <app-spinner />
} @error {
  <p>Failed to load comments.</p>
}
```

`on viewport` means the block only loads when it enters the viewport. Angular automatically code-splits the `CommentsComponent` and its dependencies into a separate chunk. Zero configuration required.

The available triggers are:

```html
@defer (on idle) { }          <!-- when browser is idle -->
@defer (on viewport) { }      <!-- when element enters viewport -->
@defer (on interaction) { }   <!-- on click or focus -->
@defer (on hover) { }         <!-- on mouse hover -->
@defer (on immediate) { }     <!-- immediately (still lazy-loaded) -->
@defer (when condition()) { } <!-- custom signal/expression -->
```

## Real world example

I used `@defer` on a dashboard page that was loading a heavy chart library. Before, the whole bundle was included upfront. After:

```html
<div class="dashboard">
  <app-summary-cards [stats]="stats()" />

  @defer (on viewport; prefetch on idle) {
    <app-analytics-chart [data]="chartData()" />
  } @placeholder {
    <div class="chart-placeholder" style="height: 400px; background: #f0f0f0;"></div>
  }
</div>
```

The `prefetch on idle` part is great — it starts downloading the chunk when the browser is idle, so by the time the user scrolls to the chart, it's probably already there. Best of both worlds between lazy and eager loading.

The initial page load time dropped noticeably. The chart library itself is around 60kb gzipped, so keeping it out of the main bundle makes a real difference.

## Migrating from the old syntax

Angular CLI provides a migration schematic:

```shell
ng generate @angular/core:control-flow
```

It converts `*ngIf`, `*ngFor` and `*ngSwitch` across your whole project. In my experience it handles 90% of cases automatically, and the remaining 10% are usually edge cases with complex template references that are easy to fix manually.

## Closing thoughts

Angular's template syntax used to be one of the weaker points of the framework compared to React's JSX or Vue's template syntax. The new control flow and `@defer` changed that completely for me.

The fact that these are compiler features — not directives from `CommonModule` — means better tree-shaking, better performance analysis by the tooling, and a lower chance of "I forgot to import something" errors.

If you're on Angular 17+ and haven't made the switch yet, the migration schematic makes it painless. Do it 🚀.

Any questions, hit me on [twitter](https://www.twitter.com/vitorpiovezam) 🤙
