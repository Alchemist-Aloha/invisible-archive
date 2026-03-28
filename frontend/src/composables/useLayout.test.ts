import { test, beforeEach } from 'node:test';
import assert from 'node:assert';
import { useLayout } from './useLayout.ts';

// Mock localStorage
let store: Record<string, string> = {};

Object.defineProperty(globalThis, 'localStorage', {
  value: {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    clear: () => {
      store = {};
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    get length() {
      return Object.keys(store).length;
    },
    key: (index: number) => Object.keys(store)[index] || null,
  },
  writable: true,
  configurable: true
});

beforeEach(() => {
  store = {};
});

test('useLayout initial state defaults to grid', () => {
  const { layoutMode } = useLayout();
  assert.strictEqual(layoutMode.value, 'grid');
});

test('useLayout initial state loads from localStorage', () => {
  localStorage.setItem('layoutMode', 'waterfall');
  const { layoutMode } = useLayout();
  assert.strictEqual(layoutMode.value, 'waterfall');
});

test('setLayoutMode updates state and localStorage', () => {
  const { layoutMode, setLayoutMode } = useLayout();
  setLayoutMode('list');
  assert.strictEqual(layoutMode.value, 'list');
  assert.strictEqual(localStorage.getItem('layoutMode'), 'list');
});

test('cycleLayout cycles through all modes', () => {
  const { layoutMode, cycleLayout } = useLayout();

  // Starts at grid
  assert.strictEqual(layoutMode.value, 'grid');

  // Cycle to list
  cycleLayout();
  assert.strictEqual(layoutMode.value, 'list');

  // Cycle to details
  cycleLayout();
  assert.strictEqual(layoutMode.value, 'details');

  // Cycle to waterfall
  cycleLayout();
  assert.strictEqual(layoutMode.value, 'waterfall');

  // Cycle back to grid
  cycleLayout();
  assert.strictEqual(layoutMode.value, 'grid');
});
