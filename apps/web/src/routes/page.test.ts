import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('home page', () => {
  it('renders the current multiplayer and Solo entry points', () => {
    render(Page);

    expect(
      screen.getByRole('heading', { level: 1, name: 'Sudoku is better together.' }),
    ).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Create a room' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Join with code' })).toBeTruthy();
    expect(screen.getByRole('link', { name: 'Play Solo' })).toBeTruthy();
  });
});
