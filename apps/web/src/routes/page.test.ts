import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('home page', () => {
  it('renders a semantic heading and an honest status', () => {
    render(Page);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Solve together. Privately.' }),
    ).toBeTruthy();
    expect(screen.getByText(/currently under construction/i)).toBeTruthy();
  });
});
