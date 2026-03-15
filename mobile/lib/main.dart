import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/settings_provider.dart';
import 'providers/explorer_provider.dart';
import 'pages/explorer_page.dart';
import 'api.dart';

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
      child: MaterialApp(
        title: 'Invisible Archive',
        theme: ThemeData(
          colorScheme: ColorScheme.fromSeed(
            seedColor: Colors.blue,
            brightness: settings.isDarkMode ? Brightness.dark : Brightness.light,
          ),
          useMaterial3: true,
          appBarTheme: const AppBarTheme(
            centerTitle: true,
          ),
          cardTheme: CardThemeData(
            clipBehavior: Clip.antiAlias,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        home: const ExplorerPage(),
        debugShowCheckedModeBanner: false,
      ),
    );
  }
}
