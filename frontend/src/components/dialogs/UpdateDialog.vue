<script setup lang="ts">
    import { computed, onMounted, onUnmounted, ref } from 'vue';
    import Button from 'primevue/button';
    import Dialog from 'primevue/dialog';
    import Message from 'primevue/message';
    import ProgressBar from 'primevue/progressbar';
    import { api } from '../../api';
    import { useI18n } from '../../i18n';
    import type { UpdateStatus } from '../../types';

    const { t } = useI18n();
    const status = ref<UpdateStatus | null>(null);
    const visible = ref(false);
    const actionPending = ref(false);
    const actionError = ref('');
    const prompted = new Set<string>();
    let pollTimer: number | undefined;
    let stopped = false;

    const state = computed(() => status.value?.state);
    const ready = computed(() => state.value === 'ready');
    const percent = computed(() => {
        const progress = status.value?.progress;
        if (!progress || progress.total <= 0) {
            return null;
        }
        return Math.min(100, Math.max(0, Math.round((progress.written / progress.total) * 100)));
    });
    const errorMessage = computed(() => actionError.value || status.value?.errorMessage || '');
    const actionLabel = computed(() => {
        if (ready.value) {
            return t('settings.update.restart');
        }
        return t(state.value === 'error' ? 'settings.update.checkNow' : 'settings.update.download');
    });
    const downloadText = computed(() =>
        percent.value === null
            ? t('settings.update.downloading')
            : t('settings.update.downloadingPercent', { percent: percent.value }),
    );

    async function refresh(): Promise<void> {
        try {
            const next = await api<UpdateStatus>('/api/update/status');
            if (stopped) {
                return;
            }
            status.value = next;
            if (!next.supported || !next.availableVersion || !['available', 'ready'].includes(next.state)) {
                return;
            }
            const key = `${next.availableVersion}:${next.state}`;
            if (!prompted.has(key)) {
                prompted.add(key);
                actionError.value = '';
                visible.value = true;
            }
        } catch {
            // Background polling must not interrupt the workspace on network failures.
        }
    }

    async function poll(): Promise<void> {
        await refresh();
        if (!stopped) {
            pollTimer = window.setTimeout(() => void poll(), state.value === 'downloading' ? 1000 : 3000);
        }
    }

    async function runAction(): Promise<void> {
        if (actionPending.value) {
            return;
        }
        const action = ready.value ? 'restart' : state.value === 'error' ? 'check' : 'download';
        actionPending.value = true;
        actionError.value = '';
        try {
            await api(`/api/update/${action}`, { method: 'POST' });
        } catch (error) {
            actionError.value = error instanceof Error ? error.message : String(error);
        } finally {
            await refresh();
            actionPending.value = false;
        }
    }

    onMounted(() => void poll());
    onUnmounted(() => {
        stopped = true;
        window.clearTimeout(pollTimer);
    });
</script>

<template>
    <Dialog
        v-model:visible="visible"
        modal
        :header="t(ready ? 'settings.update.readyTitle' : 'settings.update.promptTitle')"
        :style="{ width: '500px', maxWidth: 'calc(100vw - 32px)' }"
        class="desk-dialog update-dialog"
    >
        <div class="update-content">
            <div class="update-version">
                <i class="pi" :class="ready ? 'pi-check-circle' : 'pi-download'" aria-hidden="true" />
                <strong>{{ t('settings.update.available', { version: status?.availableVersion ?? '' }) }}</strong>
            </div>
            <p class="update-copy">
                {{ t(ready ? 'settings.update.restartHint' : 'settings.update.promptHint') }}
            </p>
            <template v-if="state === 'downloading'">
                <p class="update-copy" role="status">
                    {{ downloadText }}
                </p>
                <ProgressBar :value="percent ?? 0" :mode="percent === null ? 'indeterminate' : 'determinate'" />
            </template>
            <details v-if="status?.releaseNotes" class="update-notes">
                <summary>{{ t('settings.update.releaseNotes') }}</summary>
                <p>{{ status.releaseNotes }}</p>
            </details>
            <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
        </div>
        <template #footer>
            <Button text :label="t('settings.update.later')" @click="visible = false" />
            <Button
                v-if="state === 'available' || state === 'ready' || state === 'error'"
                :label="actionLabel"
                :icon="ready ? 'pi pi-refresh' : 'pi pi-download'"
                :loading="actionPending"
                @click="runAction"
            />
        </template>
    </Dialog>
</template>

<style scoped>
    .update-content {
        display: grid;
        gap: 16px;
    }

    .update-version {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 16px;
    }

    .update-version i {
        color: var(--accent);
        font-size: 24px;
    }

    .update-copy {
        margin: 0;
        color: var(--muted);
        line-height: 1.7;
    }

    .update-notes summary {
        cursor: pointer;
        color: var(--muted);
    }

    .update-notes p {
        max-height: 240px;
        overflow: auto;
        white-space: pre-wrap;
        overflow-wrap: anywhere;
        line-height: 1.7;
    }
</style>
