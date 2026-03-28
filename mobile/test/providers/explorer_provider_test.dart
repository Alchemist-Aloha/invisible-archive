import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:invisible_archive_mobile/api.dart';
import 'package:invisible_archive_mobile/models.dart';
import 'package:invisible_archive_mobile/providers/explorer_provider.dart';

class MockApiService extends Mock implements ApiService {}

void main() {
  late MockApiService mockApi;
  late ExplorerProvider provider;

  setUp(() {
    mockApi = MockApiService();

    // We need to provide a default mock for fetchList because ExplorerProvider
    // constructor calls fetchList(_currentPath) immediately upon instantiation.
    when(() => mockApi.fetchList(any(), sort: any(named: 'sort'), order: any(named: 'order')))
        .thenAnswer((_) async => ListResponse(items: [], effectivePath: '/'));

    provider = ExplorerProvider(mockApi);
  });

  group('ExplorerProvider search error handling', () {
    test('search sets error state when api.search throws an exception', () async {
      // Arrange
      const errorMessage = 'Search failed: 500';
      when(() => mockApi.search(any())).thenAnswer((_) async {
        // Yield to the event loop so we can observe the loading state
        await Future.delayed(const Duration(milliseconds: 10));
        throw Exception(errorMessage);
      });

      // Wait for initial fetchList to complete before testing search
      await Future.delayed(const Duration(milliseconds: 10));

      // Ensure initial state is clean
      expect(provider.error, isNull);
      expect(provider.isLoading, isFalse);

      // Act
      final searchFuture = provider.performSearch('query');

      // Verify loading state is immediately set to true
      expect(provider.isLoading, isTrue);

      await searchFuture;

      // Assert
      expect(provider.isLoading, isFalse);
      expect(provider.error, contains(errorMessage));

      verify(() => mockApi.search('query')).called(1);
    });
  });
}
