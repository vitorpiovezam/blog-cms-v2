<p align="middle">
  <img src="https://angular.dev/assets/icons/logo.svg" align="middle" title="Angular" alt="Angular logo" width="120px">
</p>

Angular 15, released in November 2022, made **Standalone Components** officially stable. The community went into hype mode. "No more NgModule! Finally!". I get the excitement, but let me give you a more honest take.

Standalone is cool. I use it. But it's not the answer to everything, and I think we need to talk about that.

## What standalone actually does

With the `standalone: true` flag, a component becomes self-sufficient. It declares its own imports and doesn't need a module to host it.

```typescript
import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatInputModule } from '@angular/material/input';

@Component({
  selector: 'app-user-card',
  standalone: true,
  imports: [CommonModule, RouterModule, ReactiveFormsModule, MatButtonModule, MatInputModule],
  template: `...`
})
export class UserCardComponent {}
```

Each component knows exactly what it needs. Looks clean in isolation. Now imagine a project with 80 components.

## The module repetition problem

Here's what nobody tells you upfront: in a real corporate application you'll end up importing the same modules over and over again in every single component.

`CommonModule` in every component that uses `*ngIf` or `*ngFor`. `ReactiveFormsModule` in every form component. Your shared UI library modules repeated across dozens of files. Your own shared service modules everywhere.

```typescript
// component-a.component.ts
imports: [CommonModule, ReactiveFormsModule, MatButtonModule, SharedModule, ...]

// component-b.component.ts
imports: [CommonModule, ReactiveFormsModule, MatButtonModule, SharedModule, ...]

// component-c.component.ts
imports: [CommonModule, ReactiveFormsModule, MatButtonModule, SharedModule, ...]

// ... and so on for 50 more components
```

That's not simplicity. That's the same boilerplate you were avoiding with NgModule, just moved from one file into every component file. You traded one centralized configuration for fifty scattered ones.

## What NgModule was actually good at

Before we throw `NgModule` in the trash, let's remember what it solved. A well-structured module system lets you group related things once:

```typescript
@NgModule({
  declarations: [UserCardComponent, UserListComponent, UserFormComponent],
  imports: [CommonModule, ReactiveFormsModule, MatButtonModule, MatInputModule],
  exports: [UserCardComponent, UserListComponent, UserFormComponent]
})
export class UserModule {}
```

Every component in that module gets access to everything imported. You change one import in one place. That's DRY. That's maintainable. That's what a team of 10 developers working on the same codebase needs.

## Where standalone actually makes sense

I'm not saying standalone is bad. I'm saying it's not the right tool for every job.

**Standalone shines for:**

- Institutional or marketing sites where you have a handful of simple pages
- Landing pages, portfolio sites, small single-feature apps
- Micro-frontends where each piece needs to be truly isolated
- Individual shared components you want to publish as a library

**NgModule still makes more sense for:**

- Large corporate applications with dozens of feature modules
- Teams where consistency and shared configuration matter
- Projects where you don't want every developer having to think about which modules to import for each new component they create

## Bootstrapping without AppModule

This part I do like, regardless of whether you use standalone everywhere:

```typescript
// main.ts
bootstrapApplication(AppComponent, {
  providers: [
    provideRouter(routes),
    provideHttpClient()
  ]
});
```

The `provide*` functions are a genuinely better API than configuring providers inside a root `AppModule`. Clean, explicit, tree-shakeable. This is a win.

## Lazy loading

`loadComponent` is also a nice addition:

```typescript
export const routes: Routes = [
  {
    path: 'dashboard',
    loadComponent: () =>
      import('./dashboard/dashboard.component').then(c => c.DashboardComponent)
  }
];
```

But `loadChildren` with a module isn't going away and is still perfectly valid, especially when you have an entire feature with multiple routes to lazy load together.

## My actual recommendation

Use standalone if you're starting something small or if the project genuinely fits the profile. Don't feel pressured to migrate a large corporate application just because the Angular team is pushing standalone as the new default.

NgModule isn't deprecated. It's not going away. It still solves real problems in real projects. The Angular team making standalone the default in the CLI doesn't mean it's the right fit for every application in existence.

Know your context and choose accordingly.
