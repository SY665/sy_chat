<template>
    <div class="flex h-screen overflow-hidden bg-white text-neutral-950 dark:bg-neutral-950 dark:text-neutral-100">
        <slot name="sidebar" />

        <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
            <slot name="header" />

            <main class="relative flex min-h-0 flex-1 flex-col overflow-hidden">
                <slot />
            </main>
        </div>
    </div>
</template>