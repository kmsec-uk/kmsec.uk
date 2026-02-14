/**
 * A single place where collections are retrieved to prevent duplicate calls to
 * getCollections
 * 
 */

import { getCollection } from "astro:content";


export const tools = await getCollection('tools')
    .then(series => series
    .sort((a,b) => b.data.priority - a.data.priority));

export const posts = await getCollection('blog')
    .then(series => series
    .sort((a,b) => b.data.date.getTime() - a.data.date.getTime()));

export const months = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December"
]

export function fmtDate(date: Date): string {
    return `${date.getDate()} ${months[date.getMonth()]} ${date.getFullYear()}`
}