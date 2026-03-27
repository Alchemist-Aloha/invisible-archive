import 'package:flutter/material.dart';

ThemeData buildAppTheme({required bool darkMode, Color? seed}) {
  final seedColor = seed ?? const Color(0xFF0066CC);
  final cs = ColorScheme.fromSeed(seedColor: seedColor, brightness: darkMode ? Brightness.dark : Brightness.light);

  return ThemeData(
    colorScheme: cs,
    useMaterial3: true,
    appBarTheme: AppBarTheme(
      centerTitle: true,
      backgroundColor: cs.surface,
      foregroundColor: cs.onSurface,
      elevation: 1,
    ),
    cardTheme: CardThemeData(
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      color: cs.surfaceVariant,
    ),
    textButtonTheme: TextButtonThemeData(
      style: TextButton.styleFrom(foregroundColor: cs.primary),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: cs.primary,
        foregroundColor: cs.onPrimary,
      ),
    ),
    filledButtonTheme: FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: cs.primaryContainer,
        foregroundColor: cs.onPrimaryContainer,
      ),
    ),
    iconTheme: IconThemeData(color: cs.onSurface),
    scaffoldBackgroundColor: cs.background,
  );
}
