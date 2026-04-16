const { LanguageClient, TransportKind } = require('vscode-languageclient/node');
const path = require('path');
const { workspace } = require('vscode');

let client;

function activate(context) {
    const serverPath = workspace.getConfiguration('riner').get('lspPath') ||
        path.join(context.extensionPath, '..', '..', '..', 'riner-lsp.exe');

    const serverOptions = {
        run: {
            command: serverPath,
            transport: TransportKind.stdio
        },
        debug: {
            command: serverPath,
            transport: TransportKind.stdio
        }
    };

    const clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'riner' }],
        synchronize: {
            fileEvents: workspace.createFileSystemWatcher('**/*.rn')
        }
    };

    client = new LanguageClient(
        'riner',
        'Riner Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
}

function deactivate() {
    if (!client) return;
    return client.stop();
}

module.exports = { activate, deactivate };