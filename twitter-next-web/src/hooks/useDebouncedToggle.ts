import { useState, useRef, useEffect, useCallback } from 'react';
import { useDebounce } from '@/hooks/useDebounce';

/**
 * A hook to manage optimistic UI state for toggle actions (Like, Follow, Retweet)
 * and debounce the actual server request to prevent spam.
 */
export function useDebouncedToggle({
    initialState,
    onMutate,
    delay = 500,
}: {
    initialState: boolean;
    onMutate: (newState: boolean) => void;
    delay?: number;
}) {
    const [optimisticState, setOptimisticState] = useState(initialState);

    // Update local state if prop changes (external revalidation)
    useEffect(() => {
        setOptimisticState(initialState);
    }, [initialState]);

    const debouncedOptimisticState = useDebounce(optimisticState, delay);
    const isMounted = useRef(false);
    const isUserAction = useRef(false);

    const onMutateRef = useRef(onMutate);

    // Update ref when onMutate changes
    useEffect(() => {
        onMutateRef.current = onMutate;
    }, [onMutate]);

    useEffect(() => {
        if (!isMounted.current) {
            isMounted.current = true;
            return;
        }

        // Only fire if the debounced state differs from the server state AND it was a user action
        if (debouncedOptimisticState !== initialState && isUserAction.current) {
            onMutateRef.current(debouncedOptimisticState);
            // Reset flag after firing
            isUserAction.current = false;
        }
    }, [debouncedOptimisticState, initialState]);

    const toggle = useCallback(() => {
        isUserAction.current = true;
        setOptimisticState((prev) => !prev);
    }, []);

    return {
        optimisticState,
        toggle,
        setOptimisticState,
    };
}
