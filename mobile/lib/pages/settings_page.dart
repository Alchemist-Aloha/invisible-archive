import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../providers/settings_provider.dart';

class SettingsPage extends StatefulWidget {
  const SettingsPage({Key? key}) : super(key: key);

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  late TextEditingController _urlController;

  @override
  void initState() {
    super.initState();
    final settings = context.read<SettingsProvider>();
    _urlController = TextEditingController(text: settings.serverUrl);
  }

  @override
  void dispose() {
    _urlController.dispose();
    super.dispose();
  }

  Future<void> _launchUrl(String url) async {
    final Uri uri = Uri.parse(url);
    if (!await launchUrl(uri, mode: LaunchMode.externalApplication)) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Could not launch $url')),
        );
      }
    }
  }

  List<Widget> _buildColorSwatches(BuildContext context) {
    final settings = context.watch<SettingsProvider>();
    final current = Color(settings.seedColor);
    final colors = <Color>[
      const Color(0xFF0066CC),
      const Color(0xFF0A9396),
      const Color(0xFF007F5F),
      const Color(0xFF9C4221),
      const Color(0xFF8B5CF6),
      const Color(0xFFEF4444),
      const Color(0xFFFFA500),
      const Color(0xFF0EA5E9),
      const Color(0xFF22C55E),
      const Color(0xFFEC4899),
    ];

    return colors.map((c) {
      final selected = c.value == current.value;
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 6.0),
        child: InkWell(
          borderRadius: BorderRadius.circular(28),
          onTap: () => context.read<SettingsProvider>().setSeedColor(c.value),
          child: Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: c,
              border: selected
                  ? Border.all(
                      color: Theme.of(context).colorScheme.onPrimaryContainer,
                      width: 3,
                    )
                  : null,
            ),
            child: selected
                ? Icon(
                    Icons.check,
                    color: Theme.of(context).colorScheme.onPrimaryContainer,
                    size: 20,
                  )
                : null,
          ),
        ),
      );
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settings'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16.0),
        children: [
          Text(
            'Server Configuration',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _urlController,
            decoration: const InputDecoration(
              labelText: 'Server URL',
              hintText: 'http://your-server:8080',
              border: OutlineInputBorder(),
              prefixIcon: Icon(Icons.dns),
            ),
            keyboardType: TextInputType.url,
          ),
          const SizedBox(height: 16),
          FilledButton.icon(
            onPressed: () async {
              final url = _urlController.text.trim();
              if (url.isNotEmpty) {
                await context.read<SettingsProvider>().setServerUrl(url);
                if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text('Settings saved'),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
                  Navigator.pop(context);
                }
              }
            },
            icon: const Icon(Icons.save),
            label: const Text('Save Server Settings'),
          ),
          const SizedBox(height: 32),
          Text(
            'Appearance',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),
          Card(
            elevation: 0,
            color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
            child: SwitchListTile(
              title: const Text('Dark Mode'),
              subtitle: const Text('Toggle dark and light themes'),
              secondary: Icon(
                context.watch<SettingsProvider>().isDarkMode
                    ? Icons.dark_mode
                    : Icons.light_mode,
              ),
              value: context.watch<SettingsProvider>().isDarkMode,
              onChanged: (val) => context.read<SettingsProvider>().setDarkMode(val),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            'Accent Color',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 56,
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: [
                  const SizedBox(width: 8),
                  ..._buildColorSwatches(context),
                  const SizedBox(width: 8),
                ],
              ),
            ),
          ),
          const SizedBox(height: 32),
          Text(
            'About',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: Theme.of(context).colorScheme.primary,
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),
          Card(
            elevation: 0,
            color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
            child: ListTile(
              leading: const Icon(Icons.code),
              title: const Text('GitHub Repository'),
              subtitle: const Text('View source code and report issues'),
              trailing: const Icon(Icons.open_in_new, size: 20),
              onTap: () => _launchUrl('https://github.com/Alchemist-Aloha/invisible-archive'),
            ),
          ),
          const SizedBox(height: 48),
          const Center(
            child: Opacity(
              opacity: 0.5,
              child: Text(
                'Invisible Archive v1.0.0',
                style: TextStyle(fontSize: 12),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
