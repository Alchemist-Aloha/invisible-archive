import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/settings_provider.dart';
import 'providers/explorer_provider.dart';
import 'pages/explorer_page.dart';
import 'api.dart';
import 'theme.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => SettingsProvider()),
        ProxyProvider<SettingsProvider, ApiService>(
          update: (_, settings, __) => ApiService.getInstance(settings.serverUrl),
        ),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final settings = context.watch<SettingsProvider>();
    final api = context.watch<ApiService>();
    
    return ChangeNotifierProvider(
      key: ValueKey(settings.serverUrl),
      create: (_) => ExplorerProvider(api),
      child: Builder(builder: (context) {
        return MaterialApp(
          title: 'Invisible Archive',
          theme: buildAppTheme(darkMode: settings.isDarkMode, seed: Color(settings.seedColor)),
          home: const ExplorerPage(),
          debugShowCheckedModeBanner: false,
        );
      }),
    );
  }
}
