import typescript from '@rollup/plugin-typescript';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import terser from '@rollup/plugin-terser';

const isDev = process.env.NODE_ENV === 'development';

export default [
  // CommonJS build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.js',
      format: 'cjs',
      sourcemap: true,
      exports: 'named'
    },
    plugins: [
      resolve({ browser: false, preferBuiltins: true }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        outputToFilesystem: true,
        declarationDir: 'dist',
        rootDir: 'src'
      }),
      !isDev && terser()
    ].filter(Boolean),
    external: ['node:fs', 'node:path', 'node:crypto']
  },
  
  // ES Module build
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.esm.js',
      format: 'esm',
      sourcemap: true,
      exports: 'named'
    },
    plugins: [
      resolve({ browser: true, preferBuiltins: false }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        declaration: false,
        declarationMap: false
      }),
      !isDev && terser()
    ].filter(Boolean),
    external: []
  },
  
  // UMD build for browsers
  {
    input: 'src/index.ts',
    output: {
      file: 'dist/index.umd.js',
      format: 'umd',
      name: 'CloudBoxSDK',
      sourcemap: true,
      exports: 'named',
      globals: {}
    },
    plugins: [
      resolve({ browser: true, preferBuiltins: false }),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        declaration: false,
        declarationMap: false
      }),
      !isDev && terser()
    ].filter(Boolean),
    external: []
  }
];