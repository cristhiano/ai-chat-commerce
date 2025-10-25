import { useState, useEffect, useCallback } from 'react';

interface UseDebounceOptions {
  delay?: number;
  leading?: boolean;
  trailing?: boolean;
}

export function useDebounce<T>(
  value: T,
  delay: number = 300,
  options: UseDebounceOptions = {}
): T {
  const { leading = false, trailing = true } = options;
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    let timeoutId: NodeJS.Timeout | null = null;

    // If leading is true, update immediately on first change
    if (leading && debouncedValue !== value) {
      setDebouncedValue(value);
    }

    // Clear existing timeout
    if (timeoutId) {
      clearTimeout(timeoutId);
    }

    // Set new timeout
    timeoutId = setTimeout(() => {
      if (trailing) {
        setDebouncedValue(value);
      }
      timeoutId = null;
    }, delay);

    // Cleanup function
    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
    };
  }, [value, delay, leading, trailing]);

  return debouncedValue;
}

// Hook for debouncing function calls
export function useDebouncedCallback<T extends (...args: any[]) => any>(
  callback: T,
  delay: number = 300,
  deps: React.DependencyList = []
): T {
  const [debounceTimer, setDebounceTimer] = useState<NodeJS.Timeout | null>(null);

  const debouncedCallback = useCallback(
    (...args: Parameters<T>) => {
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }

      const timer = setTimeout(() => {
        callback(...args);
      }, delay);

      setDebounceTimer(timer);
    },
    [callback, delay, ...deps]
  ) as T;

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }
    };
  }, [debounceTimer]);

  return debouncedCallback;
}

// Hook for debouncing async function calls
export function useDebouncedAsyncCallback<T extends (...args: any[]) => Promise<any>>(
  callback: T,
  delay: number = 300,
  deps: React.DependencyList = []
): T {
  const [debounceTimer, setDebounceTimer] = useState<NodeJS.Timeout | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const debouncedCallback = useCallback(
    async (...args: Parameters<T>) => {
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }

      const timer = setTimeout(async () => {
        setIsLoading(true);
        try {
          await callback(...args);
        } finally {
          setIsLoading(false);
        }
      }, delay);

      setDebounceTimer(timer);
    },
    [callback, delay, ...deps]
  ) as T;

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }
    };
  }, [debounceTimer]);

  return debouncedCallback;
}

// Hook for debouncing with immediate execution option
export function useDebounceWithImmediate<T>(
  value: T,
  delay: number = 300,
  immediate: boolean = false
): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);
  const [isFirstRun, setIsFirstRun] = useState(true);

  useEffect(() => {
    if (immediate && isFirstRun) {
      setDebouncedValue(value);
      setIsFirstRun(false);
      return;
    }

    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => clearTimeout(timer);
  }, [value, delay, immediate, isFirstRun]);

  return debouncedValue;
}

export default useDebounce;
