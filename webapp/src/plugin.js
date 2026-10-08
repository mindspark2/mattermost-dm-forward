import {id as pluginId} from './manifest';

const MENU_TEXT = 'Forward to…';

function startForward(postId) {
    fetch(`/plugins/${pluginId}/api/v1/action/start`, {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest',
        },
        body: JSON.stringify({post_id: postId}),
    }).catch(() => {
        // ignore network errors; user may retry
    });
}

export default class ForwardPrivatePlugin {
    initialize(registry) {
        // Always show. The server accepts channel, group, and direct-message posts.
        const filter = () => true;
        try {
            registry.registerPostDropdownMenuAction({
                text: MENU_TEXT,
                action: startForward,
                filter,
            });
        } catch (e) {
            registry.registerPostDropdownMenuAction(MENU_TEXT, startForward, filter);
        }
    }

    uninitialize() {
    }
}
